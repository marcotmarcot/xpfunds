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

EVAL = 56

def main():
  with open('get.tsv') as get:
    complete_data = []
    train_data = []
    duration = 0
    for line in get:
      array = np.array(list(map(parse, line.split('\t')[5:])), dtype=float)
      if len(array) == 0:
        continue
      for time in range(len(array)):
        array[time] = array[time] / 100 + 1
      if len(array) > duration:
        duration = len(array)
      complete_data.append(array)
      train_data.append(array[EVAL:])
  x_train = make_features(duration, train_data)
  y_train = make_ret(duration, train_data)
  x_eval = make_features(duration, complete_data)[-EVAL:]
  y_eval = make_features(duration, complete_data)[-EVAL:]
  max_train = y_train.max(axis=1).sum()
  max_eval = y_eval.max(axis=1).sum()
  x, y, w, loss = linear_regression(len(complete_data))
  optimizer = tf.train.GradientDescentOptimizer(0.1)
  train_op = optimizer.minimize(loss)
  with tf.Session() as session:
    session.run(tf.global_variables_initializer())
    feed_dict = {x: x_train, y: y_train}
    for i in range(300):
      session.run(train_op, feed_dict)
    print("loss:", (-1)*loss.eval(feed_dict)/max_train)
    print(future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), tf.truncated_normal([6, 1])).eval()/max_eval)
    print(future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), tf.zeros([6, 1])).eval()/max_eval)
    print(future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), tf.convert_to_tensor([[0.15625], [0.28125], [0.8125], [-0.46875], [0], [-0.28125]])).eval()/max_eval)
    print(future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), w).eval()/max_eval)

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
    # features /= features.max(axis=(0, 1))
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

def linear_regression(num_funds):
  x = tf.placeholder(tf.float32, shape=(None, num_funds, len(feature_functions)), name='x')
  y = tf.placeholder(tf.float32, shape=(None, num_funds, ), name='y')

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
