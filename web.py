import functools
import os

from flask import (
    Flask,
    render_template,
    redirect,
    send_from_directory,
    make_response,
)
import markdown
import sentry_sdk

from utils import load_articles


if os.getenv("SENTRY_DSN"):  # if dsn := os.getenv("xxx"); dsn != "" {} is nice in here...
    sentry_sdk.init(os.getenv("SENTRY_DSN"))


app = Flask(__name__)


articles = load_articles("./articles")


def read_article(filename):
    with open("./articles/" + filename) as f:
        title = f.readline()
        content = title + f.read()

        title = title.lstrip("#").strip()

        return title, markdown.markdown(content, extensions=["extra", "codehilite", "mdx_linkify"])


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
    return render_template("index.html", articles=articles[:32])  # magic number here...


@app.route("/all")
def all():
    return render_template("index.html", articles=articles)


@app.route("/aboutme")
@handle_exception
def aboutme():
    title, content = read_article("aboutme.md")
    return render_template("article.html", title=title, content=content)


@app.route("/interview")
@handle_exception
def interview():
    title, content = read_article("interview.md")
    return render_template("article.html", title=title, content=content)


@app.route("/friends")
@handle_exception
def friends():
    title, content = read_article("friends.md")
    return render_template("article.html", title=title, content=content)


@app.route("/articles/<filename>")
@handle_exception
def article(filename):
    if len(filename) < 6:  # `.html`
        return redirect("/404")

    filename = filename[:-5]  # remove `.html`
    title, content = read_article(filename)
    return render_template("article.html", title=title, content=content)


@app.route("/404")
def not_found():
    return render_template("404.html")


@app.route('/articles/img/<path:path>')
def serve_articles_img(path):
    return send_from_directory('articles/img', path)


@app.route('/static/<path:path>')
def serve_static(path):
    return send_from_directory('static', path)


@app.route('/favicon.ico')
def favicon():
    return send_from_directory('static', "favicon.ico")


@app.route("/rss")
def rss():
    response = make_response(render_template("rss.xml", articles=articles))
    response.headers['Content-Type'] = 'application/xml'
    return response


@app.route("/sitemap.xml")
def sitemap():
    response = make_response(render_template("sitemap.xml", articles=articles))
    response.headers['Content-Type'] = 'application/xml'
    return response
