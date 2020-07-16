"""Evaulate Sync and Async with multiprocessing."""

import sys
import argparse
import json
import multiprocessing
import logging
import asyncio
import aiofiles
from aiohttp import ClientSession
import requests

logging.basicConfig(
    format="%(asctime)s %(levelname)s:%(name)s: %(message)s",
    level=logging.INFO,
    datefmt="%H:%M:%S",
    stream=sys.stderr,
)
logger = logging.getLogger("Ticker")

class Ticker(multiprocessing.Process):
    ''' Ticker class fetches stocker ticker daily data from Yahoo. '''
    def __init__(self, task_queue, result_queue,
                 burl, uparams, odir, mode):
        multiprocessing.Process.__init__(self)
        self.task_queue = task_queue
        self.result_queue = result_queue
        self.base_url = burl
        self.params = uparams
        self.odir = odir
        self.mode = mode

    async def get(self, ticker: str, session: ClientSession) -> str:
        """Make http GET request to fetch ticker data."""
        url = f'{self.base_url}{ticker}'
        logger.debug(f'{self.name} http-get for {url}')
        resp = await session.get(url, params=self.params)
        if resp.status != 200:
            logger.error(f'{self.name}:{ticker} failed, status={resp.status}')
            return ''
        logger.debug(f'Got response [{resp.status}] for URL: {url}')
        tdata = await resp.text()
        return tdata

    async def aioprocess(self, ticker: str, session: ClientSession) -> str:
        """Issue GET for the ticker and write to file."""
        logger.debug(f'{self.name} processing_ticker {ticker}')
        fname = f'{self.odir}/{ticker}.csv'
        res = await self.get(ticker=ticker, session=session)
        if not res:
            return f'{ticker} fetch failed'
        async with aiofiles.open(fname, "a") as f:
            await f.write(res)
        return f'{ticker} fetch succeeded'

    async def asyncio_sessions(self, tickers: list) -> None:
        """Create session to concurrently fetch tickers."""
        logger.info(f'{self.name} session for {len(tickers)} tickers')
        results = []
        async with ClientSession() as session:
            tasks = []
            for t in tickers:
                tasks.append(self.aioprocess(ticker=t, session=session))
            results = await asyncio.gather(*tasks)

        # send result status
        for r in results:
            self.result_queue.put(r)

    def syncprocess(self, ticker):
        url = f'{self.base_url}{ticker}'
        logger.debug(f'{self.name} http-get for {url}')
        resp = requests.get(url, params=self.params)
        if resp.status_code != requests.codes.ok:
            logger.error(f'{self.name}:{ticker} failed, status={resp.status_code}')
            return f'{ticker} fetch failed'
        data = resp.text
        fname = f'{self.odir}/{ticker}.csv'
        with open(fname, 'w') as file:
            file.write(data)
        return f'{ticker} fetch succeeded'

    def run(self):
        pname = self.name
        tickers = []

        while True:
            t = self.task_queue.get()
            if t is None:
                logger.debug(f'{pname} Received all allocated tickers')
                break
            tickers.append(t)
            self.task_queue.task_done()

        logger.info(f'{pname} processing {self.mode} {len(tickers)} tickers')

        # Do sync or async processing
        if self.mode == "async":
            asyncio.run(self.asyncio_sessions(tickers))
        else:
            for t in tickers:
                res = self.syncprocess(t)
                self.result_queue.put(res)

        # Respond to None received in task_queue
        self.task_queue.task_done()

    def __str__(self):
        return 'Ticker %s.' % self.name

def parse_clargs():
    ''' Command line argument parser. '''
    mparser = argparse.ArgumentParser(
        description='Evaluate sync vs async multiprocessing')
    mparser.add_argument('-m',
                         '--mode',
                         action='store',
                         default='sync',
                         choices=['sync', 'async'],
                         help='evaluation mode')
    mparser.add_argument('-c',
                         '--count',
                         action='store',
                         type=int,
                         default=10,
                         help='count of tickers to fetch')
    mparser.add_argument('-p',
                         '--parallel',
                         action='store',
                         type=int,
                         default=1,
                         help='multiprocessing count')
    mparser.add_argument('-t',
                         '--tickerconf',
                         action='store',
                         type=str,
                         help='ticker config file',
                         required=True)
    mparser.add_argument('-o',
                         '--outdir',
                         action='store',
                         default='tickerdata',
                         help='output directory to store downloaded tickers')

    return mparser.parse_args()

def main():
    '''Main entry function'''

    # Parse command line arguments.
    args = parse_clargs()

    fc = args.count       # fetch count
    pc = args.parallel    # multiprocess count

    # From the ticker.conf file
    # get tickers list, base_url and url_params
    tconf = {}
    with open(args.tickerconf, 'r') as f:
        tconf = json.load(f)
    tlist = tconf['tickers']
    burl = tconf['base_url']
    uparams = tconf['params']
    logger.info(f'Processing {args.mode} {fc} tickers')

    # Task queue is used to send the tickers to processes
    # Result queue is used to get the result from processes
    tq = multiprocessing.JoinableQueue() # task queue
    rq = multiprocessing.Queue()         # result queue

    # spawning multiprocessing limited by the available cores
    if pc > multiprocessing.cpu_count():
        pc = multiprocessing.cpu_count()
    logger.info(f'Spawning {pc} gatherers...')

    tickers = [Ticker(tq, rq, burl, uparams,
                      args.outdir, args.mode) for i in range(pc)]
    for ticker in tickers:
        ticker.start()

    # enqueueing ticker jobs in task_queue
    for idx, item in enumerate(tlist):
        if idx >= fc:
            break
        tq.put(item)

    # enqueue None in task_queue to indicate completion
    for _ in range(pc):
        tq.put(None)

    tq.join()

    failc = sum(1 for i in range(fc) if 'failed' in rq.get())
    logger.info(f'Done, success: {fc-failc}/{fc}, failure: {failc}/{fc}')

if __name__ == '__main__':
    main()
