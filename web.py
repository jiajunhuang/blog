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
import requests

from utils import load_mds
from config import config

if os.getenv("SENTRY_DSN"):  # if dsn := os.getenv("xxx"); dsn != "" {} is nice in here...
    sentry_sdk.init(os.getenv("SENTRY_DSN"))

app = Flask(__name__)

articles, words = load_mds("./articles")
# title, datetime, filename, folder
all_articles = sorted(articles, key=lambda i: (i[1], i[0], i[2]), reverse=True)

SUBTITLE_MAP = {
    "golang": "Golang 教程",
    "python": "Python 教程",
    "testing": "自动化测试 教程",
}


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
def index():
    return render_template("index.html", articles=articles[:80], total_count=len(articles))  # magic number here...


@app.route("/archive")
def archive():
    return render_template("index.html", articles=articles)


@app.route("/aboutme")
@handle_exception
def aboutme():
    title, content, description = read_article("aboutme.md")
    return render_template("article.html", title=title, content=content, description=description)


@app.route("/projects")
@handle_exception
def projects():
    title, content, description = read_article("projects.md")
    return render_template("article.html", title=title, content=content, description=description)


@app.route("/friends")
@handle_exception
def friends():
    title, content, description = read_article("friends.md")
    return render_template("article.html", title=title, content=content, description=description)


@app.route("/articles/<path:filename>")
@handle_exception
def article(filename):
    recommendations = set(random.choices(all_articles, k=8))

    return render_post(filename, "article.html", read_article, recommendations=recommendations)


@app.route("/articles/<path:filename>/raw")
@handle_exception
def article_raw(filename):
    if len(filename) < 6:  # `.html`
        return redirect("/404")

    filename = filename[:-5]  # remove `.html`

    with open(os.path.join("articles", filename)) as f:
        return Response(f.read(), mimetype='text/plain')


@app.route("/jobs/<filename>")
@handle_exception
def job(filename):
    return render_post(filename, "article.html", read_job)


@app.route("/tutorial/<path:lang>/<filename>")
@handle_exception
def tutorial(lang, filename):
    subtitle = SUBTITLE_MAP.get(lang, "")

    return render_post(
        filename, "article.html", functools.partial(read_tutorial, lang), trim_html_suffix=False, subtitle=subtitle,
    )


@app.route("/404")
def not_found():
    return render_template("404.html")


@app.route("/500")
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
def rss():
    response = make_response(render_template("rss.xml", articles=all_articles))
    response.headers['Content-Type'] = 'application/xml'
    return response


@app.route("/sitemap.xml")
def sitemap():
    response = make_response(render_template("sitemap.xml", articles=all_articles))
    response.headers['Content-Type'] = 'application/xml'
    return response


@app.route("/robots.txt")
def robots():
    response = make_response(render_template("robots.txt"))
    response.headers['Content-Type'] = 'text/plain'
    return response


@app.route("/word")
def search_word():
    return redirect("https://www.google.com/search?q=site:jiajunhuang.com " + request.args.get("word", ""))


@app.route("/search", methods=["POST"])
def search():
    return redirect("https://www.google.com/search?q=site:jiajunhuang.com " + request.form.get("search"))


@app.route("/notes")
def notes():
    resp = requests.get(config.NOTES_URL).json()
    notes = resp["notes"] if resp else []
    return render_template("notes.html", title="随想", notes=notes)


@app.route("/reward")
def reward():
    user_agent = request.user_agent.string
    if "AlipayClient" in user_agent:
        return redirect(config.ALIPAY_URL)
    else:
        return redirect(config.WECHAT_PAY_URL)


if __name__ == "__main__":
    app.run("127.0.0.1", port=5000)
