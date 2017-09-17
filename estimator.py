import tensorflow as tf
import numpy as np

def model_fn(features, labels, mode):
  choice = tf.get_variable('choice', [1], dtype=tf.float64)
  y = tf.reduce_sum(choice * features['funds'] / tf.reduce_sum(choice))
  loss = tf.reduce_sum(labels - y)
  global_step = tf.train.get_global_step()
  optimizer = tf.train.GradientDescentOptimizer(0.01)
  train = tf.group(optimizer.minimize(loss), tf.assign_add(global_step, 1))
  return tf.estimator.EstimatorSpec(mode=mode, predictions=y, loss=loss, train_op=train)

def main():
  estimator = tf.estimator.Estimator(model_fn=model_fn)
  funds_train = np.array([[1.1, 1.2]])
  optimum_train = np.array([1.5])
  input_fn = tf.estimator.inputs.numpy_input_fn(
    {'funds': funds_train}, optimum_train, batch_size=1, num_epochs=None, shuffle=True)
  train_input_fn = tf.estimator.inputs.numpy_input_fn(
    {'funds': funds_train}, optimum_train, batch_size=1, num_epochs=1000, shuffle=False)

  estimator.train(input_fn=input_fn, steps=1000)
  print(estimator.evaluate(input_fn=train_input_fn))

if __name__ == '__main__':
  main()
