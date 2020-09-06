"""Using asyncio to simulate message queue"""

import os
import random
import asyncio
from abc import ABC, abstractmethod
from datetime import datetime
import aiofiles

class Consumer(ABC):
    '''Consumer/Subscriber base class'''

    @abstractmethod
    async def consume(self):
        '''consume abstract method for consuming data'''
        pass

class ConsumerFileWriter(Consumer):
    '''Consumer/Subscriber module'''
    def __init__(self, instid, msgq, rwindow):
        self._instid = instid
        self._mq = msgq
        self._rw = rwindow
        self._mcount = 0
        self._pname = "consumer_{}".format(instid)

    def get_msg(self):
        '''Consume message from queue'''
        msg = self._mq.get(self._pname)
        if msg:
            self._mcount += 1
        return msg

    async def consume(self):
        '''The process() function doing the main work of consumer'''
        procid = os.getpid()
        low, high = self._rw

        # Register with the queue
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] {self._pname} begins...')
        self._mq.register_consumer(self._pname)

        fname = "{}.txt".format(self._pname)
        fdesc = await aiofiles.open(fname, 'w')

        # work loop
        while True:
            msg = self.get_msg()
            if msg:
                await fdesc.write("{} {}\n".format(self._mcount, msg))

            # Do processing of the message
            await asyncio.sleep(random.randint(low, high))
