# Flask自动加载Blueprint

写多了 `from controllers.xxx import xxx`, `app.register_blueprint(xxx)` 就想偷懒。
于是就仿照 `unittest` 的实现思路来做了一个自动加载 `Blueprint` 的工具。使用方法如下：

```python
from loadbp import load_bp

app = Flask(__name__, template_folder="templates")
load_bp(app)
```

是不是非常简单？如果有一些 `Blueprint` 暂时还不想加载，那么设置一个属性 `_DO_NOT_LOAD_BP` 即可。例如：

```bash
$ grep _DO_NOT_LOAD_BP controllers/*
controllers/issue.py:_DO_NOT_LOAD_BP = True
controllers/user.py:_DO_NOT_LOAD_BP = True
```

下面是实现：

```python
import logging
import glob
import importlib

from flask import Blueprint, Flask

app = Flask(__name__)


def load_bp(app, path="controllers/**/*.py"):
    for file_path in glob.glob(path, recursive=True):
        module_name = file_path.split(".")[0].replace("/", ".")
        try:
            module = importlib.import_module(module_name)

            if "__init__" in file_path:
                continue

            if hasattr(module, "_DO_NOT_LOAD_BP"):
                logging.warn("ignore module %s because of attribute _DO_NOT_LOAD_BP settled", module_name)
                continue

            for attr_name in dir(module):
                attr = getattr(module, attr_name)
                if isinstance(attr, Blueprint):
                    logging.info("register %s to flask", attr_name)
                    app.register_blueprint(attr)
        except AttributeError:
            logging.error("failed to load module %s", module_name)
```

---

- https://github.com/jiajunhuang/pytemplate
- https://jiajunhuang.com/articles/2016_12_29-unittest_source_code.md.html
