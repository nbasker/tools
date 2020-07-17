# Ticker

The ticker.py is a python asyncio implementation within a multiprocessing context.

## Usage

ticker.py [-h] [-m {sync,async}] [-c COUNT] [-p PARALLEL] -t TICKERCONF [-o OUTDIR] 

## Example

Synchronous mode with one process

```
python3 ticker.py -t ticker.json -c 100 -p 1
13:54:40 INFO:Ticker: Processing sync 100 tickers
13:54:40 INFO:Ticker: Spawning 1 gatherers...
13:54:40 INFO:Ticker: Ticker-1 processing sync 100 tickers
13:55:10 INFO:Ticker: Done, success: 100/100, failure: 0/100
```

Synchronous mode with four process

```
python3 ticker.py -t ticker.json -c 300 -p 4
20:48:47 INFO:Ticker: Processing sync 300 tickers
20:48:47 INFO:Ticker: Spawning 4 gatherers...
20:48:47 INFO:Ticker: Ticker-2 processing sync 75 tickers
20:48:47 INFO:Ticker: Ticker-3 processing sync 75 tickers
20:48:47 INFO:Ticker: Ticker-1 processing sync 76 tickers
20:48:47 INFO:Ticker: Ticker-4 processing sync 74 tickers
20:49:01 INFO:Ticker: Done, success: 300/300, failure: 0/300
```

Asynchronous mode with two process

```
python3 ticker.py -t ticker.json -c 300 -p 2 -m async
20:44:22 INFO:Ticker: Processing async 300 tickers
20:44:22 INFO:Ticker: Spawning 2 gatherers...
20:44:22 INFO:Ticker: Ticker-2 processing async 135 tickers
20:44:22 INFO:Ticker: Ticker-1 processing async 165 tickers
20:44:22 INFO:Ticker: Ticker-2 session for 135 tickers
20:44:22 INFO:Ticker: Ticker-1 session for 165 tickers
20:44:24 INFO:Ticker: Done, success: 300/300, failure: 0/300
```
