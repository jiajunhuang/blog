# Flutter给Android应用签名

给Google Play交了25美元保护费，琢磨着把博客的App上架，上架的前提就是给应用签名。如下：

```bash
$ keytool -genkey -v -keystore ~/key.jks -keyalg RSA -keysize 2048 -validity 10000 -alias androidKey
```

> 记得妥善保管这个key

这个步骤里会要你填密码什么的，记住它们，下面还要用。

然后进到app源码目录编辑 `android/key.properties` 保存以下内容：

```gradle
storePassword=<你刚才填写的密码>
keyPassword=<你刚才填写的密码>
keyAlias=androidKey
storeFile=<密钥的绝对路径>
```

> 记得把这个文件加到 `.gitignore` 里：`echo 'android/key.properties' >> .gitignore`

然后编辑 `android/app/build.gradle`，把 ` android {` 替换成：

```gradle
def keystoreProperties = new Properties()
def keystorePropertiesFile = rootProject.file('key.properties')
if (keystorePropertiesFile.exists()) {
    keystoreProperties.load(new FileInputStream(keystorePropertiesFile))
}

android {
```

把

```gradle
   buildTypes {
       release {
           // TODO: Add your own signing config for the release build.
           // Signing with the debug keys for now,
           // so `flutter run --release` works.
           signingConfig signingConfigs.debug
       }
   }

```

替换成

```gradle
   signingConfigs {
       release {
           keyAlias keystoreProperties['keyAlias']
           keyPassword keystoreProperties['keyPassword']
           storeFile file(keystoreProperties['storeFile'])
           storePassword keystoreProperties['storePassword']
       }
   }
   buildTypes {
       release {
           signingConfig signingConfigs.release
       }
   }
```

最后，记得要替换一下应用的图标，我用 [这个库](https://pub.dev/packages/flutter_launcher_icons) 一键替换了。

---

我上传到Google Play之后，Google Play提示我 "You uploaded an APK or Android App Bundle that was signed in debug mode.
You need to sign your APK or Android App Bundle in release mode"，说我上传了一个debug模式签名的apk，我需要以release
模式重新打包，问题是我就是使用 `flutter build apk` 打包生成的呀，这就很奇怪了。最后解决方案是执行 `flutter clean`，
估计是因为此前我就在真机上安装过打包好的应用，而flutter为了加快打包速度缓存了一些什么。

---

参考资料：

- [给apk签名(英文)](https://flutter.dev/docs/deployment/android)
