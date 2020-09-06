"""Using asyncio to simulate message queue"""

import os
import random
import asyncio
from abc import ABC, abstractmethod
from datetime import datetime
import base_logger

logger = base_logger.getlogger(__name__)

class Producer(ABC):
    '''Producer/Publisher base class'''

    @abstractmethod
    async def produce(self):
        '''produce abstract method for producing data'''
        pass

class InMemProducer(Producer):
    '''Producer/Publisher module'''
    def __init__(self, instid, nmsgs, msgq, rwindow):
        self._instid = instid
        self._num_msgs = nmsgs
        self._mq = msgq
        self._rw = rwindow
        self._pname = "producer_{}".format(instid)

    def create_msg(self):
        '''Create message'''
        utctime = datetime.utcnow()
        utcns = int(utctime.timestamp() * 1000000)
        utcnsstr = utctime.strftime("%a, %d %b %Y %H:%M:%S.%f")
        msg = "hello from {} @ {}".format(self._pname, utcnsstr)
        return utcns, msg

    def put_msg(self, msgid, msg):
        '''Put message into queue'''
        rval = self._mq.put(msgid, msg)
        if not rval:
            logger.error("%s put message failed id: %d", self._pname, msgid)

    async def produce(self):
        '''Process function doing main work of producing messages'''
        procid = os.getpid()
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        logger.info("%s: Pid[%d] %s begins...", tstamp, procid, self._pname)
        low, high = self._rw

        for _ in range(self._num_msgs):
            # Do processing and create message
            await asyncio.sleep(random.randint(low, high))

            msgid, msg = self.create_msg()

            # Put message in the queue
            self.put_msg(msgid=msgid, msg=msg)

        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        logger.info("%s: Pid[%d] %s done, generated %d msgs",
                    tstamp, procid, self._pname, self._num_msgs)
