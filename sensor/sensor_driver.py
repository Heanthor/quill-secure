# Original file license:
# SPDX-FileCopyrightText: 2021 ladyada for Adafruit Industries
# SPDX-License-Identifier: MIT
import argparse
import sys
import time
import board
from adafruit_bme280 import basic as adafruit_bme280
import adafruit_sgp40


class SensorDriver:
    def __init__(self, poll_freq: int, debug_print: bool):
        # Create sensor object, using the board's default I2C bus.
        i2c = board.I2C()  # uses board.SCL and board.SDA
        self.bme280 = adafruit_bme280.Adafruit_BME280_I2C(i2c)
        self.sgp = adafruit_sgp40.SGP40(i2c)
        self.poll_freq = poll_freq
        self.debug_print = debug_print

        # change this to match the location's pressure (hPa) at sea level
        self.bme280.sea_level_pressure = 1017.60  # Arlington, VA

    def poll(self):
        while True:
            temperature = self.bme280.temperature
            relative_humidity = self.bme280.relative_humidity
            pressure = self.bme280.pressure
            altitude = self.bme280.altitude
            voc_index = self.sgp.measure_index(
                temperature=temperature, relative_humidity=relative_humidity)

            if self.debug_print:
                print("Temperature: %0.1f C" % temperature)
                print("Humidity: %0.1f %%" % relative_humidity)
                print("Pressure: %0.1f hPa" % pressure)
                print("Altitude = %0.2f meters" % altitude)

                print("VOC Index: ", voc_index)

            ts = int(time.time())
            sys.stdout.write(f"{ts},{temperature},{relative_humidity},{pressure},{altitude},{voc_index}")
            sys.stdout.flush()
            print()
            time.sleep(self.poll_freq)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Sensor driver')
    parser.add_argument('--poll-frequency', type=int, default=1,
                        help='poll frequency in seconds')
    parser.add_argument('--debug-print', type=bool, default=False,
                        help='poll frequency in seconds')

    args = parser.parse_args()
    driver = SensorDriver(args.poll_frequency, args.debug_print)
    driver.poll()
