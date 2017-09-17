import tensorflow as tf

def main():
  funds = tf.placeholder(tf.float32)
  choice = tf.Variable([1.0, 1.0], dtype=tf.float32)
  optimum = tf.placeholder(tf.float32)
  loss = tf.reduce_sum(optimum - choice * funds / tf.reduce_sum(choice))
  train = tf.train.GradientDescentOptimizer(0.001).minimize(loss)

  funds_train = [[1.1, 1.2]]
  optimum_train = [1.5]

  sess = tf.Session()
  sess.run(tf.global_variables_initializer())
  print(sess.run(loss, {funds: funds_train, optimum: optimum_train}))
  for i in range(100000):
    sess.run(train, {funds: funds_train, optimum: optimum_train})

  curr_choice, curr_loss = sess.run([choice, loss], {funds: funds_train, optimum: optimum_train})
  print(curr_choice, curr_choice[0]/sum(curr_choice), curr_choice[1]/sum(curr_choice), curr_loss)

if __name__ == '__main__':
  main()
