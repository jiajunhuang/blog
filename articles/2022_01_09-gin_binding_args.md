# GIN 是如何绑定参数的

在GIN这个框架里，可以通过 `Bind` 系列的函数绑定并且校验参数，我们来看看是如何实现的。

GIN 的binding分为两个系列：

- `ShouldBind` 如果参数无法通过校验，就会返回错误给调用者(对外函数均为 Bind 开头)
- `MustBind` 如果参数无法通过校验，就会自动返回400(对外函数均为 ShouldBind 开头)

如果直接调用上面两个函数，就会自动根据 `Content-Type` 猜测应该去哪里提取参数。其中 `MustBind` 底层也是调用 `ShouldBind`
来实现的：

```go
func (c *Context) MustBindWith(obj interface{}, b binding.Binding) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) // nolint: errcheck
		return err
	}
	return nil
}
```

`ShouldBind` 系列的函数有：

- `ShouldBindJSON`
- `ShouldBindXML`
- `ShouldBindYAML`
- `ShouldBindQuery`
- `ShouldBindHeader`
- `ShouldBindUri`

前面三个都是从body里读取内容，我们只看最常见的 `ShouldBindJSON`，此外我们再看一下剩下的3个。

## ShouldBindJSON

```go
// ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binding.JSON).
func (c *Context) ShouldBindJSON(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) ShouldBindWith(obj interface{}, b binding.Binding) error {
	return b.Bind(c.Request, obj)
}
```

此处调用的是 `b.Bind`，`binding.Binding` 是一个接口：

```go
// Binding describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binding interface {
	Name() string
	Bind(*http.Request, interface{}) error
}
```

因此也就是调用了 `json` 实现的 `Bind` 方法：

```go
type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	return decodeJSON(req.Body, obj)
}

func decodeJSON(r io.Reader, obj interface{}) error {
	decoder := json.NewDecoder(r)
	if EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}
```

可以看到，操作顺序就是先把 json 绑定到结构体上，然后调用 `validate` 函数，而 `validate` 最终是依靠 `Validator.ValidateStruct`
这个方法来实现参数校验的，我们来看下 `Validator.ValidateStruct` 是什么：

```go
// Validator is the default validator which implements the StructValidator
// interface. It uses https://github.com/go-playground/validator/tree/v10.6.1
// under the hood.
var Validator StructValidator = &defaultValidator{}

// StructValidator is the minimal interface which needs to be implemented in
// order for it to be used as the validator engine for ensuring the correctness
// of the request. Gin provides a default implementation for this using
// https://github.com/go-playground/validator/tree/v10.6.1.
type StructValidator interface {
	// ValidateStruct can receive any kind of type and it should never panic, even if the configuration is not right.
	// If the received type is a slice|array, the validation should be performed travel on every element.
	// If the received type is not a struct or slice|array, any validation should be skipped and nil must be returned.
	// If the received type is a struct or pointer to a struct, the validation should be performed.
	// If the struct is not valid or the validation itself fails, a descriptive error should be returned.
	// Otherwise nil must be returned.
	ValidateStruct(interface{}) error

	// Engine returns the underlying validator engine which powers the
	// StructValidator implementation.
	Engine() interface{}
}

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

// validateStruct receives struct type
func (v *defaultValidator) validateStruct(obj interface{}) error {
	v.lazyinit()
	return v.validate.Struct(obj)
}
```

可以看到 `Validator` 是符合 `StructValidator` 接口的，具体实现是 `defaultValidator`，而 `defaultValidator` 的 `ValidateStruct`
实现，底层其实是使用 `github.com/go-playground/validator/v10`，它的经典使用就是：

```go
err := validate.Struct(mystruct)
validationErrors := err.(validator.ValidationErrors)
```

JSON的绑定和校验就是这么完成的，接下来我们看看其它几种类型。

## ShouldBindQuery

```go
// ShouldBindQuery is a shortcut for c.ShouldBindWith(obj, binding.Query).
func (c *Context) ShouldBindQuery(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.Query)
}

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj interface{}) error {
	values := req.URL.Query()
	if err := mapForm(obj, values); err != nil {
		return err
	}
	return validate(obj)
}
```

从 query string 里绑定参数，就不能像JSON那样直接丢到 struct 里然后去校验了，因此我们可以看到，它这里通过 `mapForm` 函数，
把 `values` 绑定到 `obj` 之后，再去校验，我们主要来看看如何绑定：

```go
// values 的类型是 map[string][]string

func mapForm(ptr interface{}, form map[string][]string) error {
	return mapFormByTag(ptr, form, "form")
}

func mapFormByTag(ptr interface{}, form map[string][]string, tag string) error {
	// Check if ptr is a map
	ptrVal := reflect.ValueOf(ptr)
	var pointed interface{}
	if ptrVal.Kind() == reflect.Ptr {
		ptrVal = ptrVal.Elem()
		pointed = ptrVal.Interface()
	}
    // 如果ptr是一个map，并且key为字符串的时候，调用 setFormMap
	if ptrVal.Kind() == reflect.Map &&
		ptrVal.Type().Key().Kind() == reflect.String {
		if pointed != nil {
			ptr = pointed
		}
		return setFormMap(ptr, form)
	}

    // 否则调用 mappingByPtr
	return mappingByPtr(ptr, formSource(form), tag)
}

func setFormMap(ptr interface{}, form map[string][]string) error {
	el := reflect.TypeOf(ptr).Elem()

    // 如果ptr的类型符合 map[string][]string，就直接赋值过去
	if el.Kind() == reflect.Slice {
		ptrMap, ok := ptr.(map[string][]string)
		if !ok {
			return ErrConvertMapStringSlice
		}
		for k, v := range form {
			ptrMap[k] = v
		}

		return nil
	}

    // 如果ptr的类型符合 map[string]string，就把form里，value中的最后一个值赋值进去
	ptrMap, ok := ptr.(map[string]string)
	if !ok {
		return ErrConvertToMapString
	}
	for k, v := range form {
		ptrMap[k] = v[len(v)-1] // pick last
	}

	return nil
}

func mappingByPtr(ptr interface{}, setter setter, tag string) error {
	_, err := mapping(reflect.ValueOf(ptr), emptyField, setter, tag)
	return err
}

func mapping(value reflect.Value, field reflect.StructField, setter setter, tag string) (bool, error) {
	if field.Tag.Get(tag) == "-" { // just ignoring this field
		return false, nil
	}

	vKind := value.Kind()

	if vKind == reflect.Ptr {
		var isNew bool
		vPtr := value
		if value.IsNil() {
			isNew = true
			vPtr = reflect.New(value.Type().Elem())
		}
		isSet, err := mapping(vPtr.Elem(), field, setter, tag)
		if err != nil {
			return false, err
		}
		if isNew && isSet {
			value.Set(vPtr)
		}
		return isSet, nil
	}

	if vKind != reflect.Struct || !field.Anonymous {
		ok, err := tryToSetValue(value, field, setter, tag)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

    // 如果是 struct，依次设置各个field
	if vKind == reflect.Struct {
		tValue := value.Type()

		var isSet bool
		for i := 0; i < value.NumField(); i++ {
			sf := tValue.Field(i)
            // 忽略非公开字段
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			ok, err := mapping(value.Field(i), sf, setter, tag)
			if err != nil {
				return false, err
			}
			isSet = isSet || ok
		}
		return isSet, nil
	}
	return false, nil
}
```

上述方法最后通过调用 `mapping` 来设置值，如果传入的是 `struct` 或者子结构是 `struct`， 就进行递归处理。最终会通过反射，
判断是slice还是array还是普通值，尝试赋值给对应的字段。

`ShouldBindHeader` 和 `ShouldBindUri` 也是相似的逻辑。

## 总结

通过这篇文章，我们可以看到GIN是如何基于validator来实现统一的一套绑定和校验参数的机制，对于 `json`, `yaml`, `xml` 这类
值，GIN直接使用对应的库进行解析和赋值，而对于 `form`, `uri`, `query string`, `header` 等，GIN先将值提取到一个
`map[string][]string` (或者他们本身底层就已经是这个类型)里，然后再赋值给对应的 `struct`。最后，GIN使用 `validator`
进行参数校验。总体来说，GIN在这一层面的抽象还是做的很好的，除了上述类型之外，GIN还支持很多种其它类型，例如 `msgpack` 等。

希望大家有所收获。
