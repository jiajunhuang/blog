import functools
import os

from flask import (
    Flask,
    render_template,
    redirect,
    send_from_directory,
    make_response,
    request,
)
import markdown
import sentry_sdk
import requests

from utils import load_mds
from config import config

if os.getenv("SENTRY_DSN"):  # if dsn := os.getenv("xxx"); dsn != "" {} is nice in here...
    sentry_sdk.init(os.getenv("SENTRY_DSN"))

app = Flask(__name__)

articles = load_mds("./articles")
jobs = load_mds("./jobs", path="jobs")
all_articles = sorted(articles, key=lambda i: (i[1], i[0], i[2]), reverse=True)


def read_article(filename):
    return read_md("./articles", filename)


def read_job(filename):
    return read_md("./jobs", filename)


def read_md(directory, filename):
    with open(os.path.join(directory, filename)) as f:
        title = f.readline()
        content = title + f.read()

        title = title.lstrip("#").strip()

        return title, markdown.markdown(
            content, extensions=["extra", "codehilite", "mdx_linkify"],
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
    return render_template("index.html", articles=articles[:50], show_all=True)  # magic number here...


@app.route("/jobs")
def jobs_index():
    return render_template("jobs_index.html", title="招聘", articles=jobs)


@app.route("/archive")
def archive():
    return render_template("index.html", articles=articles)


def render_post(filename, template_name, load_post_func):
    if len(filename) < 6:  # `.html`
        return redirect("/404")

    filename = filename[:-5]  # remove `.html`
    title, content = load_post_func(filename)
    return render_template(template_name, title=title, content=content)


@app.route("/aboutme")
@handle_exception
def aboutme():
    title, content = read_article("aboutme.md")
    return render_template("article.html", title=title, content=content)


@app.route("/friends")
@handle_exception
def friends():
    title, content = read_article("friends.md")
    return render_template("article.html", title=title, content=content)


@app.route("/articles/<filename>")
@handle_exception
def article(filename):
    return render_post(filename, "article.html", read_article)


@app.route("/jobs/<filename>")
@handle_exception
def job(filename):
    return render_post(filename, "article.html", read_job)


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
    response = make_response(render_template("rss.xml", articles=all_articles))
    response.headers['Content-Type'] = 'application/xml'
    return response


@app.route("/sitemap.xml")
def sitemap():
    response = make_response(render_template("sitemap.xml", articles=all_articles))
    response.headers['Content-Type'] = 'application/xml'
    return response


@app.route("/search", methods=["POST"])
def search():
    return redirect("https://www.google.com/search?q=site:jiajunhuang.com " + request.form.get("search"))


@app.route("/notes")
def notes():
    resp = requests.get(config.NOTES_URL).json()
    notes = resp["notes"] if resp else []
    return render_template("notes.html", title="随想", notes=notes)


if __name__ == "__main__":
    app.run("127.0.0.1", port=5000)
