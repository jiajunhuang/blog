import logging
import functools
import os
import random

from flask import (
    Flask,
    render_template,
    redirect,
    send_from_directory,
    make_response,
    request,
    Response,
)
import markdown
import sentry_sdk
from sentry_sdk.integrations.flask import FlaskIntegration

from utils import load_mds
from config import config
from models import (
    get_session,
    Issue,
    Note,
)

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)

if config.SENTRY_DSN:
    logging.info("integrated sentry...")
    sentry_sdk.init(
        dsn=config.SENTRY_DSN,
        integrations=[FlaskIntegration()]
    )

articles, words = load_mds("./articles")
# title, datetime, filename, folder
all_articles = sorted(articles, key=lambda i: (i[1], i[0], i[2]), reverse=True)

SUBTITLE_MAP = {
    "golang": "Golang 教程",
    "python": "Python 教程",
    "testing": "自动化测试 教程",
}


@functools.lru_cache()
def get_words(top=35):
    return words.most_common(top)


# functions can be executed in jinja
app.jinja_env.globals.update(get_words=get_words)


@app.errorhandler(404)
def page_not_found(e):
    return redirect("/404")


@app.errorhandler(500)
def internal_server_error(e):
    return redirect("/500")


def read_article(filename):
    return read_md("./articles", filename)


def read_job(filename):
    return read_md("./jobs", filename)


def read_tutorial(lang, filename):
    return read_md(os.path.join("tutorial", lang), filename)


def read_md(directory, filename):
    with open(os.path.join(directory, filename)) as f:
        title = f.readline()
        body = f.read()
        content = title + body

        title = title.lstrip("#").strip()
        description = "{}...".format(body[:140])
        content = markdown.markdown(content, extensions=["extra", "tables", "codehilite", "mdx_linkify"])

        return title, content, description


def render_post(filename, template_name, load_post_func, trim_html_suffix=True, subtitle=None, recommendations=None):
    if trim_html_suffix:
        if len(filename) < 6:  # `.html`
            return redirect("/404")

        filename = filename[:-5]  # remove `.html`
    title, content, description = load_post_func(filename)
    return render_template(
        template_name,
        title=title, subtitle=subtitle, content=content, description=description, recommendations=recommendations,
    )


def handle_exception(func):
    @functools.wraps(func)
    def inner(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except FileNotFoundError:
            return redirect("/404")

    return inner


@app.route("/")
@functools.lru_cache()
def index():
    return render_template("index.html", articles=articles[:80], total_count=len(articles))  # magic number here...


@app.route("/archive")
@functools.lru_cache()
def archive():
    return render_template("index.html", articles=articles)


@app.route("/aboutme")
@handle_exception
@functools.lru_cache()
def aboutme():
    title, content, description = read_article("aboutme.md")
    return render_template("article.html", title=title, content=content, description=description)


@app.route("/natproxy")
@handle_exception
@functools.lru_cache()
def natproxy():
    title, content, description = read_article("natproxy.md")
    return render_template("article.html", title=title, content=content, description=description)


@app.route("/projects")
@handle_exception
@functools.lru_cache()
def projects():
    title, content, description = read_article("projects.md")
    return render_template("article.html", title=title, content=content, description=description)


@app.route("/friends")
@handle_exception
@functools.lru_cache()
def friends():
    title, content, description = read_article("friends.md")
    return render_template("article.html", title=title, content=content, description=description)


@app.route("/articles/<path:filename>")
@handle_exception
@functools.lru_cache()
def article(filename):
    recommendations = set(random.choices(all_articles, k=8))

    return render_post(filename, "article.html", read_article, recommendations=recommendations)


@app.route("/articles/<path:filename>/raw")
@handle_exception
@functools.lru_cache()
def article_raw(filename):
    if len(filename) < 6:  # `.html`
        return redirect("/404")

    filename = filename[:-5]  # remove `.html`

    with open(os.path.join("articles", filename)) as f:
        return Response(f.read(), mimetype='text/plain')


@app.route("/tutorial/<path:lang>/<filename>")
@handle_exception
@functools.lru_cache()
def tutorial(lang, filename):
    subtitle = SUBTITLE_MAP.get(lang, "")

    return render_post(
        filename, "article.html", functools.partial(read_tutorial, lang), trim_html_suffix=False, subtitle=subtitle,
    )


@app.route("/notes")
def notes():
    with get_session() as s:
        notes = Note.get_all(s)
        return render_template("notes.html", notes=notes)


@app.route("/sharing", defaults={'all': False})
@app.route("/sharing/<all>")
def sharing(all):
    with get_session() as s:
        issues = Issue.get_all(s) if all else Issue.get_latest_sharing(s)
        return render_template("sharing.html", issues=issues, show_all=all)


@app.route("/404")
@functools.lru_cache()
def not_found():
    return render_template("404.html")


@app.route("/500")
@functools.lru_cache()
def server_error():
    return render_template("500.html")


@app.route('/articles/img/<path:path>')
def serve_articles_img(path):
    return send_from_directory('articles/img', path)


@app.route("/tutorial/<path:lang>/img/<path:path>")
def serve_tutorial_img(lang, path):
    return send_from_directory(os.path.join("tutorial", lang, "img"), path)


@app.route('/static/<path:path>')
def serve_static(path):
    return send_from_directory('static', path)


@app.route('/favicon.ico')
def favicon():
    return send_from_directory('static', "favicon.ico")


@app.route("/rss")
@functools.lru_cache()
def rss():
    response = make_response(render_template("rss.xml", articles=all_articles))
    response.headers['Content-Type'] = 'application/xml'
    return response


@app.route("/sitemap.xml")
@functools.lru_cache()
def sitemap():
    response = make_response(render_template("sitemap.xml", articles=all_articles))
    response.headers['Content-Type'] = 'application/xml'
    return response


@app.route("/robots.txt")
@functools.lru_cache()
def robots():
    response = make_response(render_template("robots.txt"))
    response.headers['Content-Type'] = 'text/plain'
    return response


@app.route("/ads.txt")
@functools.lru_cache()
def ads():
    response = make_response(render_template("ads.txt"))
    response.headers['Content-Type'] = 'text/plain'
    return response


@app.route("/word")
def search_word():
    return redirect("https://www.google.com/search?q=site:jiajunhuang.com " + request.args.get("word", ""))


@app.route("/search", methods=["POST"])
def search():
    return redirect("https://www.google.com/search?q=site:jiajunhuang.com " + request.form.get("search"))


@app.route("/reward")
def reward():
    user_agent = request.user_agent.string
    if "MicroMessenger" in user_agent:
        return redirect(config.WECHAT_PAY_URL)
    else:
        return redirect(config.ALIPAY_URL)


if __name__ == "__main__":
    app.run("127.0.0.1", port=5000)
