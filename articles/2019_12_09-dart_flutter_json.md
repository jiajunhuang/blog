# flutter中使用RESTful接口

这篇文章简单的介绍一下，flutter中如何请求接口，并且解析响应的JSON，以及如何向服务器发送POST请求。

flutter中请求JSON的例子：

```dart
import 'dart:io';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:blogapp/consts.dart';

void checkRelease() async {
  String url = baseURL +
      "/release/latest?app=$appName&os=${Platform.operatingSystem}";

  print("check latest release with $url");
  var response = await http.get(url);
  if (response.statusCode != 200) {
    print("failed to check release: ${response.statusCode}, ${response.body}");
  } else {
    var releaseJSON = json.decode(response.body);
    print("releaseJSON: $releaseJSON");
  }
}
```

注意，简单的query string我们可以直接使用这种方式拼接，如果是复杂的，我们可以写
一个工具函数，把 `Map<String, String>` 或者 `Map<String, List<Object>>` 之类的
对象转换成query string然后再接上去。

encode或者decode JSON有这么几种用法：

- `jsonEncode(map)` 和 `json.encode(map)` 是一样的
- `jsonDecode(resp.body)` 和 `json.decode(resp.body)` 是一样的

```dart
String jsonEncode(Object object, {Object toEncodable(Object nonEncodable)}) =>
    json.encode(object, toEncodable: toEncodable);

dynamic jsonDecode(String source, {Object reviver(Object key, Object value)}) =>
    json.decode(source, reviver: reviver);
```

如果是向服务器发送JSON，那么就是：

```dart
Map<String, String> headers = {"Content-type": "application/json"};
print("request $url with $body & $headers");
var response = await http.post(url, headers: headers, body: json.encode(body));
```

注意我们要自己设置一个header。

dart的生态还是不够成熟，要是有Python中requests这样好用的库就很爽了。
