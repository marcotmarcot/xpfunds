from __future__ import absolute_import, division, print_function, unicode_literals

import pathlib
import math

import matplotlib.pyplot as plt
import pandas as pd
import seaborn as sns

import tensorflow as tf
from tensorflow import keras
from tensorflow.keras import layers

import numpy as np

def main():
  with open('get.tsv') as get:
    funds = []
    duration = 0
    for line in get:
      array = np.array(list(map(parse, line.split('\t')[5:])), dtype=float)
      if len(array) == 0:
        continue
      for time in range(len(array)):
        array[time] = array[time] / 100 + 1
      if len(array) > duration:
        duration = len(array)
      funds.append(array)
  ret = make_ret(duration, funds)
  features = make_features(duration, funds, ret)
  print(features)

def parse(field):
  return field.replace(',', '.')

def make_ret(duration, funds):
  ret = np.zeros(shape=(duration, len(funds)))
  for f in range(len(funds)):
    fund = funds[f]
    ret[-1][f] = fund[0]
    for time in range(1, len(fund)):
      ret[-1-time][f] = ret[-time][f] * fund[time]
    for time in range(1, len(fund)):
      ret[-1-time][f] **= 1. / (time + 1)
  return ret

def make_features(duration, funds, ret):
    features = np.zeros(shape=(duration, len(funds), len(feature_functions)))
    for index in range(len(feature_functions)):
      feature_functions[index](features, funds, index)
    features /= features.max(axis=(0, 1))
    return features

def set_ret(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    features[-1-(len(fund)-1)][f][index] = fund[-1]
    for time in range(len(fund)-2, -1, -1):
      features[-1-time][f][index] = features[-1-(time+1)][f][index] * fund[time]
    for time in range(0, len(fund) - 1):
      features[-1-time][f][index] **= 1. / (len(fund) - time)

def set_median(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    for time in range(len(fund)-1, -1, -1):
      s = sorted(fund[time:])
      m = s[len(s)//2]
      if len(s) % 2 == 0:
        m = (s[len(s)//2] + s[len(s)//2-1]) / 2
      features[-1-time][f][index] = m

def set_std_dev(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    total = fund[-1]
    for time in range(len(fund)-2, -1, -1):
      total += fund[time]
      count = len(fund) - time
      avg = total / count
      sumDiffs = 0.
      for i in range(time, len(fund)):
        diff = fund[time] - avg
        sumDiffs += diff * diff
      features[-1-time][f][index] = math.sqrt(sumDiffs / count)

def set_negative_month_ratio(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    negative = 0.
    nonNegative = 0.
    for time in range(len(fund)-1, -1, -1):
      if fund[time] < 1:
        negative += 1
      else:
        nonNegative += 1
      features[-1-time][f][index] = negative / (negative+nonNegative)

def set_greatest_fall(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    greatestFall = 1.
    greatestFallLen = 0.
    curr = 1.
    for time in range(len(fund)-1, -1, -1):
      curr *= fund[time]
      if fund[time] < curr:
        curr = fund[time]
      if curr < greatestFall:
        greatestFall = curr
      features[-1-time][f][index] = greatestFall

def set_greatest_fall_len(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    greatestFall = 1.
    greatestFallLen = 0.
    curr = 1.
    currLen = 0.
    for time in range(len(fund)-1, -1, -1):
      curr *= fund[time]
      currLen += 1
      if fund[time] < curr:
        curr = fund[time]
        currLen = 1
      if curr < greatestFall:
        greatestFall = curr
        greatestFallLen = currLen
      features[-1-time][f][index] = greatestFallLen

feature_functions = [
    set_ret,
    set_median,
    set_std_dev,
    set_negative_month_ratio,
    set_greatest_fall,
    set_greatest_fall_len,
]

if __name__ == '__main__':
  main()
