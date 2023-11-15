# GPT 帮我搞定了时区转换问题

前几天碰到一个时区转换问题，需要把2023年(year)第46周(week)的第2天(weekday)转换为timestamp，时区在美国洛杉矶。
其中weekday使用0表示周一，6表示周日。

我写的代码如下:

```python
import datetime
import pytz


def parse_year_week_weekday_to_timestamp(year, week, weekday, timezone):
    tzinfo = pytz.timezone(timezone)

    # weekday: Monday is 0 and Sunday is 6.
    dt = datetime.datetime.fromisocalendar(year, week, 1)
    dt += datetime.timedelta(days=weekday)

    return int(dt.replace(tzinfo=tzinfo).timestamp())


print(parse_year_week_weekday_to_timestamp(2023, 46, 1, "America/Los_Angeles"))
```

输出结果为 1699948380，但实际上结果应该为 1699948800。把问题告诉GPT，GPT给的代码大概如下：

```python
import datetime
import pytz


def parse_year_week_week_day_to_timestamp_by_pytz(year, week, weekday, timezone):
    tzinfo = pytz.timezone(timezone)

    dt = datetime.datetime.fromisocalendar(year, week, 1)
    dt += datetime.timedelta(days=weekday)

    return int(tzinfo.localize(dt).timestamp())


print(parse_year_week_week_day_to_timestamp_by_pytz(2023, 46, 1, "America/Los_Angeles"))
```

可以得到正确结果。这两种实现方式最大的区别在哪里呢？一个在于直接调用了 `datetime.datetime.replace` 函数，另一个则是
使用了 `pytz.localize` 函数。因此我翻到源码看了一下区别，发现 `replace` 函数并没有处理夏令时冬令时的问题，而是直接
拿着给的 tzinfo 构建一个新的 datetime 对象，`localize` 函数处理了夏令时冬令时的问题。

这是 `datetime.datetime.replace` 函数的代码：

```c
static PyObject *
datetime_replace(PyDateTime_DateTime *self, PyObject *args, PyObject *kw)
{
    PyObject *clone;
    PyObject *tuple;
    int y = GET_YEAR(self);
    int m = GET_MONTH(self);
    int d = GET_DAY(self);
    int hh = DATE_GET_HOUR(self);
    int mm = DATE_GET_MINUTE(self);
    int ss = DATE_GET_SECOND(self);
    int us = DATE_GET_MICROSECOND(self);
    PyObject *tzinfo = HASTZINFO(self) ? self->tzinfo : Py_None;
    int fold = DATE_GET_FOLD(self);

    if (! PyArg_ParseTupleAndKeywords(args, kw, "|iiiiiiiO$i:replace",
                                      datetime_kws,
                                      &y, &m, &d, &hh, &mm, &ss, &us,
                                      &tzinfo, &fold))
        return NULL;
    if (fold != 0 && fold != 1) {
        PyErr_SetString(PyExc_ValueError,
                        "fold must be either 0 or 1");
        return NULL;
    }
    tuple = Py_BuildValue("iiiiiiiO", y, m, d, hh, mm, ss, us, tzinfo);
    if (tuple == NULL)
        return NULL;
    clone = datetime_new(Py_TYPE(self), tuple, NULL);
    if (clone != NULL) {
        DATE_SET_FOLD(clone, fold);
    }
    Py_DECREF(tuple);
    return clone;
}
```

可以看到，`replace` 函数直接拿给定的时区构建一个新的对象，而没有处理时区转换的问题。

下面看 `localize` 的实现：

```python
    def localize(self, dt, is_dst=False):
        '''Convert naive time to local time.

        This method should be used to construct localtimes, rather
        than passing a tzinfo argument to a datetime constructor.

        is_dst is used to determine the correct timezone in the ambigous
        period at the end of daylight saving time.

        >>> from pytz import timezone
        >>> fmt = '%Y-%m-%d %H:%M:%S %Z (%z)'
        >>> amdam = timezone('Europe/Amsterdam')
        >>> dt  = datetime(2004, 10, 31, 2, 0, 0)
        >>> loc_dt1 = amdam.localize(dt, is_dst=True)
        >>> loc_dt2 = amdam.localize(dt, is_dst=False)
        >>> loc_dt1.strftime(fmt)
        '2004-10-31 02:00:00 CEST (+0200)'
        >>> loc_dt2.strftime(fmt)
        '2004-10-31 02:00:00 CET (+0100)'
        >>> str(loc_dt2 - loc_dt1)
        '1:00:00'

        Use is_dst=None to raise an AmbiguousTimeError for ambiguous
        times at the end of daylight saving time

        >>> try:
        ...     loc_dt1 = amdam.localize(dt, is_dst=None)
        ... except AmbiguousTimeError:
        ...     print('Ambiguous')
        Ambiguous

        is_dst defaults to False

        >>> amdam.localize(dt) == amdam.localize(dt, False)
        True

        is_dst is also used to determine the correct timezone in the
        wallclock times jumped over at the start of daylight saving time.

        >>> pacific = timezone('US/Pacific')
        >>> dt = datetime(2008, 3, 9, 2, 0, 0)
        >>> ploc_dt1 = pacific.localize(dt, is_dst=True)
        >>> ploc_dt2 = pacific.localize(dt, is_dst=False)
        >>> ploc_dt1.strftime(fmt)
        '2008-03-09 02:00:00 PDT (-0700)'
        >>> ploc_dt2.strftime(fmt)
        '2008-03-09 02:00:00 PST (-0800)'
        >>> str(ploc_dt2 - ploc_dt1)
        '1:00:00'

        Use is_dst=None to raise a NonExistentTimeError for these skipped
        times.

        >>> try:
        ...     loc_dt1 = pacific.localize(dt, is_dst=None)
        ... except NonExistentTimeError:
        ...     print('Non-existent')
        Non-existent
        '''
        if dt.tzinfo is not None:
            raise ValueError('Not naive datetime (tzinfo is already set)')

        # Find the two best possibilities.
        possible_loc_dt = set()
        for delta in [timedelta(days=-1), timedelta(days=1)]:
            loc_dt = dt + delta
            idx = max(0, bisect_right(
                self._utc_transition_times, loc_dt) - 1)
            inf = self._transition_info[idx]
            tzinfo = self._tzinfos[inf]
            loc_dt = tzinfo.normalize(dt.replace(tzinfo=tzinfo))
            if loc_dt.replace(tzinfo=None) == dt:
                possible_loc_dt.add(loc_dt)

        if len(possible_loc_dt) == 1:
            return possible_loc_dt.pop()

        # If there are no possibly correct timezones, we are attempting
        # to convert a time that never happened - the time period jumped
        # during the start-of-DST transition period.
        if len(possible_loc_dt) == 0:
            # If we refuse to guess, raise an exception.
            if is_dst is None:
                raise NonExistentTimeError(dt)

            # If we are forcing the pre-DST side of the DST transition, we
            # obtain the correct timezone by winding the clock forward a few
            # hours.
            elif is_dst:
                return self.localize(
                    dt + timedelta(hours=6), is_dst=True) - timedelta(hours=6)

            # If we are forcing the post-DST side of the DST transition, we
            # obtain the correct timezone by winding the clock back.
            else:
                return self.localize(
                    dt - timedelta(hours=6),
                    is_dst=False) + timedelta(hours=6)

        # If we get this far, we have multiple possible timezones - this
        # is an ambiguous case occurring during the end-of-DST transition.

        # If told to be strict, raise an exception since we have an
        # ambiguous case
        if is_dst is None:
            raise AmbiguousTimeError(dt)

        # Filter out the possiblilities that don't match the requested
        # is_dst
        filtered_possible_loc_dt = [
            p for p in possible_loc_dt if bool(p.tzinfo._dst) == is_dst
        ]

        # Hopefully we only have one possibility left. Return it.
        if len(filtered_possible_loc_dt) == 1:
            return filtered_possible_loc_dt[0]

        if len(filtered_possible_loc_dt) == 0:
            filtered_possible_loc_dt = list(possible_loc_dt)

        # If we get this far, we have in a wierd timezone transition
        # where the clocks have been wound back but is_dst is the same
        # in both (eg. Europe/Warsaw 1915 when they switched to CET).
        # At this point, we just have to guess unless we allow more
        # hints to be passed in (such as the UTC offset or abbreviation),
        # but that is just getting silly.
        #
        # Choose the earliest (by UTC) applicable timezone if is_dst=True
        # Choose the latest (by UTC) applicable timezone if is_dst=False
        # i.e., behave like end-of-DST transition
        dates = {}  # utc -> local
        for local_dt in filtered_possible_loc_dt:
            utc_time = (
                local_dt.replace(tzinfo=None) - local_dt.tzinfo._utcoffset)
            assert utc_time not in dates
            dates[utc_time] = local_dt
        return dates[[min, max][not is_dst](dates)]
```

这里就比较复杂了，做了一堆的猜测和转换，主要是处理夏令时冬令时的问题，因此python中解决时区问题，还是要避免
直接使用replace函数。

## 时区问题涉及的概念

常见的概念有这么几个：

- GMT（Greenwich Mean Time）， 格林威治平时（也称格林威治时间）。它规定太阳每天经过位于英国伦敦郊区的皇家格林威治天文台的时间为中午12点。它曾经是世界时间标准，但是现在不再使用了，因为精度不够被UTC替代了。
- UTC（Coodinated Universal Time），协调世界时，又称世界统一时间、世界标准时间、国际协调时间。由于英文（CUT）和法文（TUC）的缩写不同，作为妥协，简称UTC。UTC 是现在全球通用的时间标准，全球各地都同意将各自的时间进行同步协调。UTC 时间是经过平均太阳时（以格林威治时间GMT为准）、地轴运动修正后的新时标以及以秒为单位的国际原子时所综合精算而成。
- DST（Daylight Saving Time），夏令时又称夏季时间，或者夏时制。是为了充分使用夏季白天阳光提出的一种概念，中国没有使用夏令时，因此这个概念我们可能比较陌生。使用夏令时的国家，会把时间往前调一小时。
- 本地时间：我们平时在哪里使用，哪里的时间就是我们的本地时间。一般本地时间会对应UTC时间加上一些偏移量，比如中国采用北京时间，在东八区，因此北京时间就是 UTC+8，我们平时跑代码时，如果不单独配置时区，那么默认就是使用本地时间。
- UNIX TIMESTAMP：UNIX时间戳，这个是与时区无关的，他是一种数字类型，是从1970年1月1日00:00:00到现在的秒数，是一个时间的增量。

弄清楚了这些，我们就大概知道时间的转换关系了，例如如果我们要将本地时间(北京时间，UTC+8)，转换成洛杉矶时间（UTC-8）：

```python
def beijing_time_to_los_angeles_time(dt):
    return dt.astimezone(pytz.timezone("America/Los_Angeles"))


print(beijing_time_to_los_angeles_time(datetime.datetime.now()))
```
