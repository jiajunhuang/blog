# 自己封装一个好用的Dart HTTP库

用过Python的同学，大概都用过requests这个库，这么好用的库，就会想其他语言有没有这个库。Dart没有，所以自己封装：

```dart
import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:blogapp/consts.dart';

String genQueryString(Map<String, dynamic> queryStringMap) {
  if (queryStringMap == null) {
    return "";
  }

  List<String> queryStringList = [];
  queryStringMap.forEach((k, v) {
    if (v is String) {
      queryStringList.add("$k=$v");
    } else if (v is List) {
      if (v != null) {
        v.forEach((vv) => queryStringList.add("$k=$vv"));
      }
    } else {
      print("bad value type, ignore value $k-$v");
    }
  });
  return queryStringList.join("&");
}

Future<http.Response> postJSON(String uri, Object jsonBody,
    {Map<String, dynamic> queryString, Map<String, String> headers}) async {
  final qs = genQueryString(queryString);
  String url = "$baseURL$uri?$qs";
  if (headers == null) {
    headers = Map<String, String>();
  }
  headers["Content-type"] = "application/json";
  headers["User-Agent"] = userAgent;

  return http.post(url, headers: headers, body: json.encode(jsonBody));
}

Future<http.Response> getJSON(String uri,
    {Map<String, dynamic> queryString, Map<String, String> headers}) async {
  final qs = genQueryString(queryString);
  String url = "$baseURL$uri?$qs";
  if (headers == null) {
    headers = Map<String, String>();
  }
  headers["Content-type"] = "application/json";
  headers["User-Agent"] = userAgent;

  return http.get(url);
}
```

自己封装一个，我感觉还挺好用。
