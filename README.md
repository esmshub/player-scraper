# player-scraper
A utility for generating a CSV export of all players for a particular ESMS game.

### Supported games

- [SSL](http://www.ssl2001.ukhome.net)
- [FFO](https://ffomanager.com)

## Usage

### SSL
```
SSL Player Scraper
------------------
Usage of ./bin/debug/ssl_scraper:
  -dir string
        Local directory of rosters (default ".")
  -download-files
        Download rosters from the SSL website if missing (default true)
  -max-concurrent int
        Number of concurrent requests when loading rosters (default 5)
  -output-dir string
        Output directory for CSV files (default ".")
  -stop-on-error
        Stop all requests on first error (default true)
  -teams-url string
        Teams page URL on SSL website (default "http://www.ssl2001.ukhome.net/teams.htm")
```
### FFO
```
FFO Player Scraper
------------------
Usage of ffo_scraper:
  -clubs-url string
        Clubs page URL on FFO website (default "https://www.ffomanager.com/clubs.html")
  -dir string
        Local directory of rosters (default ".")
  -download-files
        Download rosters from the FFO website if missing (default true)
  -max-concurrent int
        Number of concurrent requests when loading rosters (default 5)
  -output-dir string
        Output directory for CSV files (default ".")
  -stop-on-error
        Stop all requests on first error (default true)
```