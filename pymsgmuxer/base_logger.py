import logging

def getlogger(name=__name__, level=logging.INFO):
    '''getlogger to get the logger for the module'''
    logger = logging.getLogger(name)
    logger.setLevel(level)

    # create console handler and set level to debug
    console = logging.StreamHandler()
    # create formatter
    formatter = logging.Formatter('%(asctime)s %(name)-12s %(levelname)-8s %(message)s')
    # add formatter to console
    console.setFormatter(formatter)
    # add console to logger
    logger.addHandler(console)
    return logger
