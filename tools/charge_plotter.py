import argparse
import csv
import datetime
import math

from matplotlib import ticker
from matplotlib.pyplot import figure
import matplotlib.pyplot as plt
from matplotlib.ticker import AutoMinorLocator


def main():
    argparser = argparse.ArgumentParser()
    argparser.set_defaults(cmd=lambda: None, cmd_args=lambda x: [])
    argparser.add_argument('path', type=str, metavar="PATH")

    args = argparser.parse_args()

    csvfile = open(args.path, 'r', newline='')
    wr = csv.reader(csvfile,delimiter=',')

    date_series = []
    voltage_series = []
    current_series = []

    min_voltage = 500
    max_voltage = 0
    min_current = 500
    max_current = 0


    first_date = None
    entries_i = iter(wr)
    next(entries_i)
    for entry in entries_i:
        dt = datetime.datetime.strptime(entry[0], "%Y-%m-%d %H:%M:%S")
        if first_date is None:
            first_date = dt

        time_sec = (dt - first_date).total_seconds() / 60 / 60
        date_series.append(time_sec)

        voltage = float(entry[1])
        current = float(entry[2])


        voltage_series.append(voltage)
        current_series.append(current)


        min_voltage = math.floor(min(min_voltage, voltage))
        max_voltage = math.ceil(max(max_voltage, voltage))
        min_current = math.floor(min(min_current, current))
        max_current = math.ceil(max(max_current, current))


    v_scale_mult = 1

    fig, (ax1_current) = plt.subplots(1, 1)
    fig.set_size_inches(8.5, 4)

    ax1_current.grid(True)
    ax1_current.set_xlabel("Time [h]", color='black')
    ax1_voltage = ax1_current.twinx()

    # Voltage
    ax1_voltage.plot(date_series, voltage_series, color='green')
    ax1_voltage.yaxis.set_major_formatter(ticker.FormatStrFormatter('%g V'))
    ax1_voltage.set_ylim([min_voltage * v_scale_mult, max_voltage * v_scale_mult])
    ax1_voltage.set_ylabel("Voltage", color='green')

    # Current
    ax1_current.plot(date_series, current_series, color='red')
    ax1_current.yaxis.set_major_formatter(ticker.FormatStrFormatter('%d A'))
    ax1_current.set_ylim([min_current, max_current])
    ax1_current.set_ylabel("Current", color='red')
    ax1_current.set_yticks(range(0, max_current + 1, 2))

    plt.show()
    
main()