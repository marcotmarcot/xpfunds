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
import scipy

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
  x_batch = make_features(duration, funds)
  y_batch = make_ret(duration, funds)
  max = y_batch.max(axis=1).sum()
  x, y, w, loss = linear_regression(duration, len(funds))
  optimizer = tf.train.GradientDescentOptimizer(0.1)
  train_op = optimizer.minimize(loss)
  with tf.Session() as session:
    session.run(tf.global_variables_initializer())
    print(future_return(tf.convert_to_tensor(x_batch, dtype=tf.float32), tf.convert_to_tensor(y_batch, dtype=tf.float32), tf.truncated_normal([6, 1])).eval())
    print(future_return(tf.convert_to_tensor(x_batch, dtype=tf.float32), tf.convert_to_tensor(y_batch, dtype=tf.float32), tf.zeros([6, 1])).eval())
    feed_dict = {x: x_batch, y: y_batch}
    for i in range(30):
      session.run(train_op, feed_dict)
      print(i, "loss:", (-1)*loss.eval(feed_dict)/max, w.eval())


def parse(field):
  return field.replace(',', '.')

def make_ret(duration, funds):
  ret = np.zeros(shape=(duration, len(funds)))
  for f in range(len(funds)):
    fund = funds[f]
    for time in range(len(fund)):
      ret[-1-time][f] = np.prod(fund[:time+1]) ** (1. / (time+1))
  return ret

def make_features(duration, funds):
    features = np.zeros(shape=(duration, len(funds), len(feature_functions)))
    for index in range(len(feature_functions)):
      feature_functions[index](features, funds, index)
    features = (features - np.mean(features, axis=(0, 1))) / np.std(features, axis=(0, 1))
    return features

def set_ret(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    for time in range(len(fund)):
      features[-1-time][f][index] = np.prod(fund[time:]) ** (1. / (len(fund)-time))

def set_median(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    for time in range(len(fund)):
      features[-1-time][f][index] = np.median(fund[time:])

def set_std_dev(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    for time in range(len(fund)):
      features[-1-time][f][index] = np.std(fund[time:])

def set_negative_month_ratio(features, funds, index):
  duration = len(features)
  for f in range(len(funds)):
    fund = funds[f]
    for time in range(len(fund)):
      period = fund[time:]
      features[-1-time][f][index] = len(np.extract(period < 1, period)) / len(period)

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
      features[-1-time][f][index] = greatestFallLen / (len(fund) - time)

feature_functions = [
    set_ret,
    set_median,
    set_std_dev,
    set_negative_month_ratio,
    set_greatest_fall,
    set_greatest_fall_len,
]

def linear_regression(duration, num_funds):
  x = tf.placeholder(tf.float32, shape=(duration, num_funds, len(feature_functions)), name='x')
  y = tf.placeholder(tf.float32, shape=(duration, num_funds, ), name='y')

  with tf.variable_scope('lreg') as scope:
    w = tf.Variable(tf.truncated_normal((len(feature_functions), 1)), name='W')
    ret = future_return(x, y, w)

  return x, y, w, -ret

def future_return(x, y, w):
  stock_score = tf.squeeze(tf.einsum("ijk,kl->ijl", x, w))
  frac = tf.math.softmax(stock_score, axis=1)
  return tf.reduce_sum(tf.multiply(frac, y))

if __name__ == '__main__':
  main()
