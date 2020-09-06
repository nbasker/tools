"""Using asyncio to simulate message queue"""

import os
import random
import asyncio
import base_logger
import mq
import producer
import consumer

logger = base_logger.getlogger(__name__)

async def aio_sessions(loop):
    '''Create session to concurrently fetch tickers.'''
    procid = os.getpid()
    logger.info("Pid[%d]: asyncio session...", procid)

    # Create message queue, producers and consumers
    mqueue = mq.MsgQueue(retention=30, cleanfreq=10)
    pro1 = producer.InMemProducer(instid=1, nmsgs=25, msgq=mqueue, rwindow=[2, 3])
    pro2 = producer.InMemProducer(instid=2, nmsgs=15, msgq=mqueue, rwindow=[1, 5])
    pro3 = producer.InMemProducer(instid=3, nmsgs=5, msgq=mqueue, rwindow=[1, 2])
    con1 = consumer.ConsumerFileWriter(instid=1, msgq=mqueue, rwindow=[1, 4])
    con2 = consumer.ConsumerFileWriter(instid=2, msgq=mqueue, rwindow=[2, 3])
    con3 = consumer.ConsumerFileWriter(instid=3, msgq=mqueue, rwindow=[2, 4])

    # Shuffle to start in random order
    funcs = []
    funcs.append(pro1.produce())
    funcs.append(pro2.produce())
    funcs.append(con1.consume())
    funcs.append(con2.consume())
    funcs.append(mqueue.cleanqueue())
    random.shuffle(funcs)

    # Create asyncio tasks for execution
    tasks = []
    producer_tasks = []
    for func in funcs:
        await asyncio.sleep(random.randint(1, 5))
        if "Producer" in func.__qualname__:
            producer_tasks.append(loop.create_task(func))
        else:
            tasks.append(loop.create_task(func))

    # Start later consumer and producer respectively
    await asyncio.sleep(35)
    tasks.append(loop.create_task(con3.consume()))
    await asyncio.sleep(10)
    producer_tasks.append(loop.create_task(pro3.produce()))

    # Wait for all producers to complete
    await asyncio.gather(*producer_tasks)

    # Cancel consumers and message queue clean task
    logger.info("Producers completed, give few minutes for consumers")
    await asyncio.sleep(120)
    for task in tasks:
        task.cancel()
    try:
        await asyncio.gather(*tasks)
    except asyncio.CancelledError:
        logger.info("Consumer cancelled now")

def main():
    '''Main entry function'''
    loop = asyncio.get_event_loop()
    loop.run_until_complete(aio_sessions(loop))

if __name__ == '__main__':
    main()
