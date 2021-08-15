import json
import re
import time
import logging
import aiohttp
import asyncio
from typing import (
    List,
    Dict,
    Any
)

log = logging.getLogger("monitorlog")


class RequestTracer():
    def __init__(self) -> None:
        # traceinfo : ti
        self._ti: Dict[str, Any] = {}
        # routetracer : rt
        self._rt = aiohttp.TraceConfig()
        self._rt.on_request_start.append(self._req_start)
        self._rt.on_request_end.append(self._req_end)

    @property
    def trace_config(self) -> aiohttp.TraceConfig:
        return self._rt

    def clear_trace(self) -> None:
        self._ti.clear()

    def get_rtt(self, url: str) -> float:
        rtt = 0.0
        if url in self._ti:
            if 'start_time' in self._ti[url] and 'end_time' in self._ti[url]:
                total = self._ti[url]['end_time'] - self._ti[url]['start_time']
                rtt = round(total * 1000, 2)
        return rtt

    async def _req_start(self, session, context, params):
        self._ti[str(params.url)] = {
            'start_time': asyncio.get_event_loop().time()
        }

    async def _req_end(self, session, context, params):
        self._ti[str(params.url)]['end_time'] = asyncio.get_event_loop().time()


class AsyncSiteMonitor:
    def __init__(self, moncfg: Dict[str, Any]) -> None:
        self._custname = moncfg['custname']
        self._sitelist = [site['url'] for site in moncfg['sites']
                          if 'url' in site]
        self._pollfreq = moncfg['pollfreq']
        self._retrydelay = moncfg['retrydelay']
        self._redict = {s['url']: re.compile(s['regex'])
                        for s in moncfg['sites'] if 'url' in s}

    async def _fetch(self,
                     session: aiohttp.ClientSession,
                     url: str) -> Dict[str, Any]:
        mets: Dict[str, Any] = {}
        try:
            log.info(f"metric fetch url : {url}")
            async with session.get(url) as resp:
                mets['url'] = url
                mets['status'] = resp.status
                mets['reason'] = resp.reason
                mets['method'] = resp.method
                mets['content_type'] = resp.content_type
                mets['title'] = 'Not Found'
                checkRE = self._redict[url]
                async for line in resp.content:
                    match = checkRE.search(line.decode("utf-8"))
                    if match:
                        mets["title"] = match.group(1)
                        break
        except (aiohttp.ClientError, asyncio.TimeoutError):
            emsg = f"ClientError in http call to {url}"
            log.error(emsg)
            raise ConnectionError(emsg)
        return mets

    async def collect(self, rt: RequestTracer,
                      client: aiohttp.ClientSession) -> List[Dict[str, Any]]:
        rt.clear_trace()
        tasks = [self._fetch(client, s) for s in self._sitelist]
        metrics = []
        for future in asyncio.as_completed(tasks):
            try:
                metric = await future
                if not metric:
                    continue
                metric['rtt'] = rt.get_rtt(metric['url'])
                metric['customer'] = self._custname
                metric['timestamp'] = int(time.time()*1000)
                metrics.append(metric)
                pkey = f"{self._custname}/{metric['url']}"
            except ConnectionError as e:
                log.error(f"Exception in collection {repr(e)}")
        return metrics

    async def collectloop(self) -> None:
        retry_count = 0
        rt = RequestTracer()
        while True:
            try:
                async with aiohttp.ClientSession(
                        trace_configs=[rt.trace_config]) as client:
                    while True:
                        log.info("Starting metric collection loop")
                        await self.collect(rt, client)
                        await asyncio.sleep(self._pollfreq)
            except (aiohttp.ClientError, asyncio.TimeoutError):
                retry_count += 1
                log.info("aiohttp.ClientSession error")
                await asyncio.sleep(self._retrydelay)
                log.info(f"retrying after {retry_count} attempts")

    async def monitor(self) -> None:
        try:
            await self.collectloop()
        except Exception as e:
            log.error(f"Exception triggered {repr(e)}, args : {e.args}")

    async def shutdown(self, signal) -> None:
        log.info(f'Received exit signal {signal.name}...')
        tasks = [
            t for t in asyncio.all_tasks() if t is not asyncio.current_task()
        ]
        [task.cancel() for task in tasks]
        log.info(f'Cancelling {len(tasks)} outstanding tasks')
        await asyncio.gather(*tasks, return_exceptions=True)
        asyncio.get_event_loop().stop()
        log.info('Stopped loop')
