# 自己动手写一个k8s controller

如果要处理一个云原生业务，尤其是跨云业务，k8s controller 是无可避免的，这篇博客就记录我自己折腾学习 k8s controller，从
最开始简单的照着 `sample-controller` 来，到扩展成一个支持多任务、多步骤的 controller。

## 定义 CRD

自定义的 controller 一般主要是为了处理 CRD，CRD 简单来说，就是一堆yaml，用来扩展自定义资源，比如 k8s 官方提供了 `Deployment`，
那么如果我想自己搞一个 `webapps` 对象呢？

所以有这么一个 CRD:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: webapps.jiajunhuang.com
spec:
  group: jiajunhuang.com
  names:
    kind: WebApp
    singular: webapp
    plural: webapps
    shortNames:
    - webapp
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        required: ["spec"]
        properties:
          spec:
            type: object
            required: ["version", "replicas"]
            properties:
              version:
                type: string
                description: The version of the webapp
                minLength: 1
              replicas:
                type: integer
                minimum: 0
                description: The number of replicas of the webapp
              image:
                type: string
                description: The container image to use
                minLength: 1
          status:
            type: object
            properties:
              availableReplicas:
                type: integer
                description: The number of available replicas
              conditions:
                type: array
                items:
                  type: object
                  properties:
                    lastTransitionTime:
                      type: string
                      format: date-time
                    phase:
                      type: string
                    reason:
                      type: string
                    message:
                      type: string
    subresources:
      status: {}
    additionalPrinterColumns:
    - name: Replicas
      type: integer
      description: The number of replicas
      jsonPath: .spec.replicas
    - name: Version
      type: string
      description: The version of the webapp
      jsonPath: .spec.version
    - name: Age
      type: date
      jsonPath: .metadata.creationTimestamp
```

## 组织 controller

有了CRD以后，根据CRD还需要组织一下代码，大概目录如下：

```bash
$ tree .
pkg
$ tree .
.
├── crd
│   └── crd_webapps_jiajunhuang_com.yaml
├── go.mod
├── go.sum
├── hack
│   ├── boilerplate.go.txt
│   ├── tools.go
│   └── update-codegen.sh
├── pkg
│   ├── apis
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── jiajunhuang.com
│   │       └── v1
│   │           ├── doc.go
│   │           ├── register.go
│   │           ├── types.go
```

- `doc.go` 内容为：

```go
// +k8s:deepcopy-gen=package
// +groupName=jiajunhuang.com

// Package v1 is the v1 version of the API.
package v1

```

- `register.go` 内容为：

```go
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{Group: "jiajunhuang.com", Version: "v1"}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&WebApp{},
		&WebAppList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
```

- `types.go` 的内容为：

```go
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebApp 是一个自定义资源示例
type WebApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebAppSpec   `json:"spec"`
	Status WebAppStatus `json:"status,omitempty"`
}

// WebAppSpec 定义了 WebApp 的期望状态
type WebAppSpec struct {
	// 在这里添加你的规格字段
	Image    string `json:"image"`
	Version  string `json:"version"`
	Replicas int32  `json:"replicas"`
}

// WebAppStatus 定义了 WebApp 的实际状态
type WebAppStatus struct {
	// 在这里添加你的状态字段
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebAppList 包含 WebApp 的列表
type WebAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WebApp `json:"items"`
}
```

然后执行 `./hack/update-codegen.sh`

```bash
$ ./hack/update-codegen.sh 
Generating deepcopy code for 1 targets
Generating client code for 1 targets
Generating lister code for 1 targets
Generating informer code for 1 targets
```

这里我踩了一个坑，一开始不知道为什么总是报错：

```bash
./hack/update-codegen.sh
Generating deepcopy code for 1 targets
F0216 11:17:21.869640 1489301 main.go:107] Error: failed making a parser: error(s) in "github.com/jiajunhuang/test/pkg/apis/jiajunhuang.com/v1":
-: module github.com/jiajunhuang/test@latest found (v0.0.0-00010101000000-000000000000, replaced by ./), but does not contain package github.com/jiajunhuang/test/pkg/apis/jiajunhuang.com/v1
```

后来把 `pkg/apis` 独立拆分成一个新的模块，才不报错了，花了很久排查，但是最终原因我也没搞懂。

## 编写 controller

此处大概就是参考 `sample-controller` 来写，代码详见：https://github.com/jiajunhuang/test-k8s-controller/commit/1924a052d69cb974917e2294c3b05f3191895e1f#diff-243ebed2765f75e6a54f57167212fefb08c3b2a85967ad2acbc0eb78919019c1

接下来就是一顿折腾，改造成了最终版本，支持多任务、多步骤，以下是运行时的日志：

```bash
I0216 15:08:31.963311 1716546 task1.go:20] "执行 PreCreate" webapp="webapp-sample" step="step1"
I0216 15:08:31.963325 1716546 task1.go:25] "执行 Create" webapp="webapp-sample" step="step1"
I0216 15:08:31.963330 1716546 task1.go:30] "执行 PostCreate" webapp="webapp-sample" step="step1"
I0216 15:08:31.964083 1716546 task2.go:20] "执行 PreCreate" webapp="webapp-sample" step="step2"
I0216 15:08:31.964092 1716546 task2.go:25] "执行 Create" webapp="webapp-sample" step="step2"
I0216 15:08:31.964097 1716546 task2.go:30] "执行 PostCreate" webapp="webapp-sample" step="step2"
I0216 15:08:31.964868 1716546 task3.go:20] "执行 PreCreate" webapp="webapp-sample" step="step3"
I0216 15:08:31.964875 1716546 task3.go:25] "执行 Create" webapp="webapp-sample" step="step3"
I0216 15:08:31.964879 1716546 task3.go:30] "执行 PostCreate" webapp="webapp-sample" step="step3"
I0216 15:08:31.965583 1716546 task1.go:20] "执行 PreCreate" webapp="webapp-sample" step="step1"
I0216 15:08:31.965590 1716546 task1.go:25] "执行 Create" webapp="webapp-sample" step="step1"
I0216 15:08:31.965594 1716546 task1.go:30] "执行 PostCreate" webapp="webapp-sample" step="step1"
I0216 15:08:31.966385 1716546 task2.go:20] "执行 PreCreate" webapp="webapp-sample" step="step2"
I0216 15:08:31.966394 1716546 task2.go:25] "执行 Create" webapp="webapp-sample" step="step2"
I0216 15:08:31.966400 1716546 task2.go:30] "执行 PostCreate" webapp="webapp-sample" step="step2"
I0216 15:08:31.967157 1716546 task3.go:20] "执行 PreCreate" webapp="webapp-sample" step="step3"
I0216 15:08:31.967165 1716546 task3.go:25] "执行 Create" webapp="webapp-sample" step="step3"
I0216 15:08:31.967169 1716546 task3.go:30] "执行 PostCreate" webapp="webapp-sample" step="step3"
I0216 15:08:34.701611 1716546 task3.go:35] "执行 PreDelete" webapp="webapp-sample" step="step3"
I0216 15:08:34.701622 1716546 task3.go:40] "执行 Delete" webapp="webapp-sample" step="step3"
I0216 15:08:34.701629 1716546 task3.go:45] "执行 PostDelete" webapp="webapp-sample" step="step3"
I0216 15:08:34.701637 1716546 controller.go:215] "成功完成删除操作" webapp="webapp-sample" executor="*tasks.Step3TaskExecutor"
I0216 15:08:34.702551 1716546 task2.go:35] "执行 PreDelete" webapp="webapp-sample" step="step2"
I0216 15:08:34.702561 1716546 task2.go:40] "执行 Delete" webapp="webapp-sample" step="step2"
I0216 15:08:34.702566 1716546 task2.go:45] "执行 PostDelete" webapp="webapp-sample" step="step2"
I0216 15:08:34.702574 1716546 controller.go:215] "成功完成删除操作" webapp="webapp-sample" executor="*tasks.Step2TaskExecutor"
I0216 15:08:34.703391 1716546 task1.go:35] "执行 PreDelete" webapp="webapp-sample" step="step1"
I0216 15:08:34.703401 1716546 task1.go:40] "执行 Delete" webapp="webapp-sample" step="step1"
I0216 15:08:34.703407 1716546 task1.go:45] "执行 PostDelete" webapp="webapp-sample" step="step1"
I0216 15:08:34.706199 1716546 controller.go:215] "成功完成删除操作" webapp="webapp-sample" executor="*tasks.Step1TaskExecutor"
I0216 15:08:34.706223 1716546 controller.go:151] "WebApp 已经被删除" webapp="default/webapp-sample"
```

代码比较多，直接见 github: https://github.com/jiajunhuang/test-k8s-controller

核心结构：

- main 函数中传入多个任务

```go
controller, err := controller.NewController(ctx, kubeClient, exampleClient,
		exampleInformerFactory.Jiajunhuang().V1().WebApps(),
		[]tasks.TaskExecutor{
			&tasks.Step1TaskExecutor{},
			&tasks.Step2TaskExecutor{},
			&tasks.Step3TaskExecutor{},
		})
```

每个 `task` 都符合接口：

```go
// TaskExecutor 定义任务执行器接口
type TaskExecutor interface {
	Name() string
	PreCreate(ctx context.Context, webapp *v1.WebApp) error
	Create(ctx context.Context, webapp *v1.WebApp) error
	PostCreate(ctx context.Context, webapp *v1.WebApp) error
	PreDelete(ctx context.Context, webapp *v1.WebApp) error
	Delete(ctx context.Context, webapp *v1.WebApp) error
	PostDelete(ctx context.Context, webapp *v1.WebApp) error
}
```

对于所有的task，创建时，会按照 `PreCreate -> Create -> PostCreate` 的顺序来执行，删除时，会按照逆序来执行 `PreDelete -> Delete -> PostDelete`，
核心代码：

```go
func (c *Controller) syncHandler(ctx context.Context, objectRef types.NamespacedName) error {
	logger := klog.FromContext(ctx)

	webapp, err := c.webappsLister.WebApps(objectRef.Namespace).Get(objectRef.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	// 如果对象不存在，直接返回
	if apierrors.IsNotFound(err) {
		logger.Info("WebApp 已经被删除", "webapp", objectRef.String())
		return nil
	}

	// 处理删除操作
	if webapp.DeletionTimestamp != nil {
		// execute deletion tasks in reverse order
		for i := len(c.tasks) - 1; i >= 0; i-- {
			if err := c.handleDeletion(ctx, webapp, c.tasks[i]); err != nil {
				return err
			}
		}
		return nil
	}

	// 处理创建/更新操作
	for _, executor := range c.tasks {
		if err := c.handleCreation(ctx, webapp, executor); err != nil {
			return err
		}
	}
	return nil
}

// 处理删除操作
func (c *Controller) handleDeletion(ctx context.Context, webapp *v1.WebApp, executor tasks.TaskExecutor) error {
	logger := klog.FromContext(ctx)

	// 获取最新版本的对象
	currentWebApp, err := c.webclientset.JiajunhuangV1().WebApps(webapp.Namespace).Get(ctx, webapp.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("获取最新版本的 WebApp 失败: %v", err)
	}

	// 检查是否包含我们的 finalizer
	if !containsFinalizer(currentWebApp.Finalizers, webappFinalizer) {
		return nil
	}

	// 执行删除相关操作
	if err := executor.PreDelete(ctx, webapp); err != nil {
		return fmt.Errorf("执行 PreDelete 失败: %v", err)
	}
	if err := executor.Delete(ctx, webapp); err != nil {
		return fmt.Errorf("执行 Delete 失败: %v", err)
	}
	if err := executor.PostDelete(ctx, webapp); err != nil {
		return fmt.Errorf("执行 PostDelete 失败: %v", err)
	}

	// 只有当这是最后一个任务时才移除 finalizer
	if executor == c.tasks[0] {
		// 使用最新版本的对象移除 finalizer
		webappCopy := currentWebApp.DeepCopy()
		webappCopy.Finalizers = removeFinalizer(webappCopy.Finalizers, webappFinalizer)
		_, err = c.webclientset.JiajunhuangV1().WebApps(webapp.Namespace).Update(ctx, webappCopy, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("移除 finalizer 失败: %v", err)
		}
	}

	logger.Info("成功完成删除操作", "webapp", webapp.Name, "executor", fmt.Sprintf("%T", executor))
	return nil
}

// 处理创建/更新操作
func (c *Controller) handleCreation(ctx context.Context, webapp *v1.WebApp, executor tasks.TaskExecutor) error {
	logger := klog.FromContext(ctx)

	// 添加 finalizer 前先获取最新版本的对象
	currentWebApp, err := c.webclientset.JiajunhuangV1().WebApps(webapp.Namespace).Get(ctx, webapp.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取最新版本的 WebApp 失败: %v", err)
	}

	// 确保有 finalizer
	if !containsFinalizer(currentWebApp.Finalizers, webappFinalizer) {
		webappCopy := currentWebApp.DeepCopy()
		webappCopy.Finalizers = append(webappCopy.Finalizers, webappFinalizer)
		currentWebApp, err = c.webclientset.JiajunhuangV1().WebApps(webapp.Namespace).Update(ctx, webappCopy, metav1.UpdateOptions{})
		if err != nil {
			if apierrors.IsInvalid(err) {
				logger.Error(err, "资源验证失败",
					"webapp", klog.KObj(webapp),
					"details", err.(*apierrors.StatusError).ErrStatus.Details)
				return nil
			}
			return fmt.Errorf("添加 finalizer 失败: %v", err)
		}
		// 更新成功后，使用最新版本继续处理
		webapp = currentWebApp
	}

	if err := executor.PreCreate(ctx, webapp); err != nil {
		return fmt.Errorf("执行 PreCreate 失败: %v", err)
	}
	if err := executor.Create(ctx, webapp); err != nil {
		return fmt.Errorf("执行 Create 失败: %v", err)
	}
	if err := executor.PostCreate(ctx, webapp); err != nil {
		return fmt.Errorf("执行 PostCreate 失败: %v", err)
	}

	return nil
}
```

完整代码就不贴了，建议直接去 Github 看。

---

参考资料：

- https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/
- https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/
- https://github.com/kubernetes/sample-controller
