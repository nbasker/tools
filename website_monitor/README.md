# Website monitor and metrics collector

A monitor that periodically checks the accessibility of a given website and collects metrics such as round-trip-time, response code, time of access.

## References

* [Python Asyncio Series](https://bbc.github.io/cloudfit-public-docs/)
* [AsyncIO for the Working Python Developer](https://medium.com/@yeraydiazdiaz/asyncio-for-the-working-python-developer-5c468e6e2e8e)
* [AIOHttp](https://docs.aiohttp.org/en/stable/index.html)
* [Making a million requests with Python AIOHttp](https://pawelmhm.github.io/asyncio/python/aiohttp/2016/04/22/asyncio-aiohttp.html)
* [Python Type Checking Guide](https://realpython.com/python-type-checking/)
* [Pytest for AsyncIO](https://pypi.org/project/pytest-asyncio/)

## Setup

The monitoring utility takes in a configuration file and also provides lint (falkes8) checking and unit testing as described below.

### Config

The config/configtemplate.json is the file through which parameters can be passed to control the monitoring service.
The user can make a copy and provide specific credentials and URL's to monitor.

### Flakes

The **make flake** command does a flake8 check (pyflakes, PEP8 and McCabe complexity checker).

### Unit Testing and CI
The **make test** and **make ci** commands are used to run unit test and CI.

### Invoking Monitor
* **python3 metric_collector.py** starts the monitor of a set of URL's specified in config file.

## Design Considerations

1. Using AsyncIO approach as the application is IO intensive. Enables high utilization of CPU core through concurrency.
2. Reuse of aiohttp.ClientSession object over multiple connections. Each session can hold 100 connections and hence can monitor 100 websites.
3. Streaming Response Content rather than reading the entire content for parsing and matching regex.
4. Used AsyncIO Client Tracing feature to mark the round-trip-time.
5. Regex matching is flexible that can be customised for each URL that is monitored.

## Further Enhancements

1. Monitoring of a website happens at "pollfreq" intervals which is a configuration parameter. Need to add variance so that the monitor can schedule http client requests by dividing the websites into groups. This enables a single instance to monitor a high number of sites. Not all sites are monitored at the same interval period.
2. Taking advantage of connection-pool-size per client-session and limiting connections per host. This can improve efficiency to pack more websites to monitor. Additionally make the following parameters configurable through config-file.
    1. Website client timeout value
    1. Number of client sessions
    1. Number of connection pools per client session
    1. Number of connections per host
2. Only GET method is implemented, can be enhanced to do PUT, POST as well.
3. The AsyncIO Client Tracing needs to be enhanced to capture fine granular statistics including in case of exceptions.
4. The regex is only one per URL it can be made into a list. Additionally, the extraction of match is not very versatile. Needs further implementation.
5. Just two unit test cases implemented. Need to use pytest-fixture and mock to cover more cases.
5. Packaging the monitor service in docker container and use k8s deployments so that secrets can be managed and auto-scaling can be implemented.
