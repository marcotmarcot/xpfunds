import matplotlib.pyplot as plt
import numpy as np
import tensorflow as tf
from tensorflow import keras
import time

EPOCHS = 500

def main():
  train_data = np.loadtxt('train_data.tsv', ndmin=2)
  train_labels = np.loadtxt('train_labels.tsv')
  test_data = np.loadtxt('test_data.tsv', ndmin=2)
  test_labels = np.loadtxt('test_labels.tsv')

  # Remove lines with zero
  # train_data = train_data[:, np.all(train_data != 0, axis=0)]
  # test_data = test_data[:, np.all(test_data != 0, axis=0)]

  # Normalize
  mean = train_data.mean(axis=0)
  std = train_data.std(axis=0)
  train_data = (train_data - mean) / std
  test_data = (test_data - mean) / std

  # Model
  model = keras.Sequential([
      keras.layers.Dense(
          64, activation=tf.nn.relu, input_shape=(train_data.shape[1],)),
      keras.layers.Dense(64, activation=tf.nn.relu),
      keras.layers.Dense(1)
  ])
  optimizer = tf.train.RMSPropOptimizer(0.001)
  model.compile(loss='mse', optimizer=optimizer, metrics=['mae'])

  # Stop
  early_stop = keras.callbacks.EarlyStopping(monitor='val_loss', patience=20)

  # Train
  history = model.fit(
      train_data,
      train_labels,
      epochs=EPOCHS,
      validation_split=0.2,
      verbose=0)

  [loss, mae] = model.evaluate(test_data, test_labels, verbose=0)
  print("Testing set Mean Abs Error: {}".format(mae))
  test_predictions = model.predict(test_data).flatten()
  np.savetxt('test_predictions.tsv', test_predictions, delimiter='\t')


class Fund:
  def __init__(self, line):
    fields = line.strip().split('\t')
    monthly = []
    for f in fields[5:]:
      monthly.append(1 + float(f.replace(',', '.')) / 100)
    self.duration = len(self.raw)
    # The monthly profit of the fund. Starts from the last month we have data
    # and goes back to the first month we have data.
    self.monthly = np.array(monthly)
    self.duration = len(self.monthly)

  def annual(self, end, start):
    return self.monthly[end:start].prod() ** (1.0/((start - end)/12.0))


def data_and_labels(funds, end, start):
  data = []
  labels = []
  for f in funds:
    for time in range(end + 1, start - 1):
      if time >= f.duration:
        break
      duration = f.duration - time
      negative_months = 0
      largest_drop = 999.0
      for monthly in f.monthly[time:f.duration]:
        if monthly < 1.0:
          negative_months += 1
        drop = monthly
        for s in range(e + 1, f.duration):
          drop *= f.raw[s]
          if drop < largest_drop:
            largest_drop = drop
      data.append([
          f.annual(time, f.duration),
          duration,
          negative_months / duration,
          np.array(values).std(axis=0),
          largest_drop
      ])
      labels.append(f.annual(end, time))
  return np.array(data), np.array(labels)


def plot_history(history):
  plt.figure()
  plt.xlabel('Epoch')
  plt.ylabel('Mean Abs Error')
  plt.plot(
      history.epoch,
      np.array(history.history['mean_absolute_error']),
      label='Train Loss')
  plt.plot(history.epoch, np.array(history.history['val_mean_absolute_error']), label='Val loss')
  plt.legend()
  plt.ylim([0, 5])
  plt.show()


if __name__ == '__main__':
  main()
