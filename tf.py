import matplotlib.pyplot as plt
import numpy as np
import tensorflow as tf
from tensorflow import keras
import time

EPOCHS = 500

def main():
  train_data_list = []
  train_labels_list = []
  for i in range(100):
    train_data_list.append([i])
    train_labels_list.append(i + i)
  print(train_data_list)
  print(train_labels_list)
  train_data = np.array(train_data_list)
  train_labels = np.array(train_labels_list)
  test_data = np.array([[4], [5], [6]])
  test_labels = np.array([8, 10, 12])

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
      verbose=0,
      callbacks=[PrintDot()])

  [loss, mae] = model.evaluate(test_data, test_labels, verbose=0)
  print("Testing set Mean Abs Error: {}".format(mae * 1000))

  test_predictions = model.predict(test_data).flatten()
  print(test_predictions)


class PrintDot(keras.callbacks.Callback):
  def on_epoch_end(self, epoch, logs):
    if epoch % 100 == 0: print('')
    print('.', end='')


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
