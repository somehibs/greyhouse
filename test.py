import RPi.GPIO as GPIO
import time
GPIO.setmode(GPIO.BCM)
pin=17
GPIO.setup(pin, GPIO.OUT)
GPIO.output(pin, 1)
time.sleep(1)
GPIO.cleanup()
