# Let's Watch

Library and tool for interacting with various film sites and APIs.

Pulls together APIs from:

* Letterboxd - uses go-letterboxd to scrape film data
* TMDB - Uses the legit TMDB API

Caches results to a local Redis cache. For quickly accessing data.

## Prerequisites

This project uses the amazing Cobra/Viper library for command line parsing. To bootstrap with your personal info, use something like this in your `~/.letswatch.yaml`:

```yaml
---
letterboxd-username: 'my-username'
subscribed-to:
  - "HBO Max"
  - "Shudder"
```

## Examples

By default, the films we have already watched (and logged on letterboxd.com) are
excluded from any recommended listings.

Get a list of the top 250 narrative films that are available on my streaming services

```shell
$ letswatch recommend --list dave/official-top-250-narrative-feature-films --only-my-streaming
   • Getting lists
   • Getting watched films
   • Fetching list films       slug=official-top-250-narrative-feature-films username=dave
- title: Harakiri
  release_year: 1962
  imdb_link: https://www.imdb.com/title/tt0056058
  language: ja
  runtime: 2h15m0s
  streaming_on:
    - Criterion Channel
  genres:
    - Action
    - Drama
    - History
... SNIP ...
   • Run stats                 duration=10.685447243s total_items=133
...
```
