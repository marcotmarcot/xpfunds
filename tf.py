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
NUM_FUNDS = 5

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
  x_eval = np.concatenate((np.zeros(shape=(duration-EVAL, len(complete_data), len(feature_functions))), make_features(duration, complete_data)[-EVAL:]))
  y_eval = np.concatenate((np.zeros(shape=(duration-EVAL, len(complete_data))), make_ret(duration, complete_data)[-EVAL:]))
  max_train = y_train.max(axis=1).sum()
  max_eval = y_eval.max(axis=1).sum()
  x, y, w, loss = linear_regression(duration, len(complete_data))
  optimizer = tf.compat.v1.train.GradientDescentOptimizer(0.1)
  train_op = optimizer.minimize(loss)
  with tf.compat.v1.Session() as session:
    session.run(tf.compat.v1.global_variables_initializer())
    feed_dict = {x: x_train, y: y_train}
    for i in range(300):
      session.run(train_op, feed_dict)
    print("train:", (-1)*loss.eval(feed_dict)/max_train)
    print("random_train:", future_return(tf.convert_to_tensor(x_train, dtype=tf.float32), tf.convert_to_tensor(y_train, dtype=tf.float32), tf.truncated_normal([len(feature_functions)*NUM_FUNDS, 1])).eval()/max_train)
    print("random_eval:", future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), tf.truncated_normal([len(feature_functions)*NUM_FUNDS, 1])).eval()/max_eval)
    print("ones:", future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), tf.ones([len(feature_functions)*NUM_FUNDS, 1])).eval()/max_eval)
    print("zeros:", future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), tf.zeros([len(feature_functions)*NUM_FUNDS, 1])).eval()/max_eval)
    print("eval:", future_return(tf.convert_to_tensor(x_eval, dtype=tf.float32), tf.convert_to_tensor(y_eval, dtype=tf.float32), w).eval()/max_eval)

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

def linear_regression(duration, num_funds):
  x = tf.compat.v1.placeholder(tf.float32, shape=(duration, num_funds, len(feature_functions)), name='x')
  y = tf.compat.v1.placeholder(tf.float32, shape=(duration, num_funds, ), name='y')

  with tf.compat.v1.variable_scope('lreg') as scope:
    w = tf.Variable(tf.truncated_normal((len(feature_functions)*NUM_FUNDS, 1)), name='W')
    ret = future_return(x, y, w)

  return x, y, w, -ret

def future_return(x, y, w):
  total = 0
  for i in range(NUM_FUNDS):
    stock_score = tf.squeeze(tf.einsum("ijk,kl->ijl", x, w[i*len(feature_functions):(i+1)*len(feature_functions)]))
    # mask = tf.argsort(stock_score, direction='DESCENDING') < 1
    # frac = tf.divide(tf.where(tf.math.logical_not(mask), tf.where(mask, stock_score, tf.zeros_like(stock_score)), tf.ones_like(stock_score)), NUM_FUNDS)
    frac = tf.divide(tf.math.softmax(stock_score, 1), NUM_FUNDS)
    total += tf.reduce_sum(tf.multiply(frac, y))
  return total

if __name__ == '__main__':
  main()
