import datetime
import contextlib

from sqlalchemy import (
    create_engine,
    Column,
    DateTime,
    Integer,
    String,
)
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

from config import config

engine = create_engine(config.SQLALCHEMY_DB_URI, echo=config.SQLALCHEMY_ECHO)
Session = sessionmaker(bind=engine)

Base = declarative_base()


class BaseMixin:
    id = Column(Integer, primary_key=True, autoincrement=True)

    created_at = Column(DateTime, nullable=False, default=datetime.datetime.now)
    updated_at = Column(DateTime, nullable=False, default=datetime.datetime.now, onupdate=datetime.datetime.now, index=True)  # noqa
    deleted_at = Column(DateTime, nullable=True, index=True)

    @classmethod
    def get_all(cls, session, limit=None, order_by=None):
        q = session.query(cls)
        if limit:
            q = q.limit(limit)
        if order_by:
            q = q.order_by(order_by)
        else:
            q = q.order_by(cls.id.desc())

        return q.all()

    @classmethod
    def get_by_id(cls, session, item_id):
        return session.query(cls).filter(
            cls.id == item_id,
            cls.deleted_at.is_(None),
        ).first()

    @classmethod
    def get_latest_one(cls, session):
        return session.query(cls).order_by(cls.id.desc()).first()


@contextlib.contextmanager
def get_session():
    s = Session()
    try:
        yield s
        s.commit()
    except Exception:
        s.rollback()
        raise
    finally:
        s.close()


class Issue(Base, BaseMixin):
    __tablename__ = "issue"

    content = Column(String, nullable=False, default="", doc="内容")
    url = Column(String, default="", doc="链接")

    @classmethod
    def get_latest_sharing(cls, session, limit=30):
        return session.query(cls).order_by(cls.id.desc()).limit(limit).all()


class Note(Base, BaseMixin):
    __tablename__ = "note"

    content = Column(String, nullable=False, doc="内容")
