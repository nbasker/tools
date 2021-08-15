import signal
import argparse
import json
import logging
import asyncio
from monitor import monitor


log = logging.getLogger("monitorlog")


def main():
    '''Main entry function'''

    ch = logging.StreamHandler()
    ch.setLevel(logging.INFO)
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(message)s')
    ch.setFormatter(formatter)
    log.addHandler(ch)
    log.setLevel(logging.INFO)
    log.info("Starting website monitor producer")

    parser = argparse.ArgumentParser(
            description='Website availability metrics collector')
    parser.add_argument("-c", "--config", type=str,
                        default='./configs/config.json',
                        help="configuration file")

    args = parser.parse_args()
    with open(args.config, 'r') as f:
        config = json.load(f)

    asm = monitor.AsyncSiteMonitor(config['monitor'], None)
    loop = asyncio.get_event_loop()

    signals = (signal.SIGHUP, signal.SIGTERM, signal.SIGINT)
    for s in signals:
        loop.add_signal_handler(
            s, lambda s=s: asyncio.create_task(asm.shutdown(s)))

    try:
        loop.run_until_complete(asm.monitor())
    except asyncio.exceptions.CancelledError:
        # The shutdown tasks need to run to completion
        loop.run_forever()
    finally:
        loop.close()
    log.info('Stopped monitor after cleaning up...')


if __name__ == '__main__':
    main()
