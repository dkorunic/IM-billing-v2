# IM-billing-v2

[![GitHub license](https://img.shields.io/github/license/dkorunic/IM-billing-v2.svg)](https://github.com/dkorunic/IM-billing-v2/blob/master/LICENSE.txt)
[![GitHub release](https://img.shields.io/github/release/dkorunic/IM-billing-v2.svg)](https://github.com/dkorunic/IM-billing-v2/releases/latest)
[![codebeat badge](https://codebeat.co/badges/97692d96-db24-40dc-8fda-a9b5be1eb09c)](https://codebeat.co/projects/github-com-dkorunic-im-billing-v2-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/dkorunic/IM-billing-v2)](https://goreportcard.com/report/github.com/dkorunic/IM-billing-v2)

## About

IM-billing-v2 is a simple Google calendar based tracking/billing system. When you
have a set of tasks performed in your Google calendar with each of the
entries belonging to a specific "sub"-calendar, you can easily print out
those for a specific (and any) time frame, sum them and make a simple
billing calculation.

## API

- Google Calendar API: [https://developers.google.com/calendar/v3/reference/](https://developers.google.com/calendar/v3/reference/)

## Installation

There are two ways of installing IM-billing-v2:

### Manual

Download your preferred flavor from [the releases](https://github.com/dkorunic/IM-billing-v2/releases/latest) page and install manually.

### Using go get

```shell
go install github.com/dkorunic/IM-billing-v2@latest
```

## Usage

```shell
NAME
  IM-billing-v2

FLAGS
  -c, --calendar STRING    calendar name
  -s, --start STRING       start date (YYYY-MM-DD)
  -e, --end STRING         end date (YYYY-MM-DD)
  -x, --search STRING      search string (substring match in event description)
      --config STRING      config file (optional)
  -t, --timeout DURATION   Google Calendar API timeout (default: 1m0s)
  -h, --help               display help
  -d, --dash               use dashes when printing totals
  -r, --recurring          include recurring events
```

Typical use example to fetch calendar items in your primary calendar from `01/01/2017` to `01/01/2018` and sum only calendar events prefixed with `CLIENT:` prefix:

```shell
./IM-billing-v2 \
  --search CLIENT: \
  --start 2017-01-01 \
  --end 2018-08-01
```
