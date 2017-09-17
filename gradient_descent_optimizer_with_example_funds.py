# Possible strategies
#
# 1. Select: Train with the raw fund data as input and the selection of the best
# fund as output, with an array where only the best fund is 1 and the others are
# 0.
#   * Disadvantages:
#     * Ignores most of the data on the future.
#
# 2. Loss: Use the loss function to train, looking only at the future.
#   * Disadvantages:
#     * Ignores the data of the past.
#     * Generates strategy that always chose the same funds
#
# 3. Predict: Predict the rentability of the future based on the rentability of
# the past.
import tensorflow as tf

class Fund:
  def __init__(self, line):
    fields = line.strip().split('\t')
    self.name = fields[0]
    self.prof = []
    for f in fields[4:]:
      self.prof.insert(0, 1 + float(f.replace(',', '.')) / 100)
    self.table = [[]] * (len(self.prof) + 1)
    for start in range(len(self.prof), 0, -1):
      self.table[start] = [0] * start
      self.table[start][start - 1] = self.prof[start - 1]
      for end in range(start - 2, -1, -1):
        self.table[start][end] = self.table[start][end + 1] * self.prof[end]

  def time(self):
    return len(self.prof)

  def profitability(self, start, end):
    if start > len(self.prof):
      return 0
    return self.table[start][end]


def annual(value, months):
  return value ** (1.0/(months/12.0))


def choice_loss(funds, duration, optima, choice):
  loss = 0
  loss_time = 0
  for i in range(0, len(funds)):
    prof = 0
    ratio = 0
    for j in range(0, len(choice)):
      prof += choice[j] * funds[i][j]
      ratio += choice[j]
    prof = prof / ratio
    if prof == 0:
      continue
    loss += (annual(max(optima[i]), duration[i]) - annual(prof, duration[i])) * duration[i]
    loss_time += duration[i]
  return 100.0 * loss / loss_time


def main():
  funds_data = []
  choice_init = []
  max_time = 0
  with open("get.tsv") as f:
    for line in f:
      fund = Fund(line)
      funds_data.append(fund)
      choice_init.append(1.0)
      if fund.time() > max_time:
        max_time = fund.time()

  funds_train = []
  duration_train = []
  optima_train = []
  for start in range(1, max_time + 1):
    fund_train = []
    optimum = 0
    for fund in funds_data:
      p = fund.profitability(start, 0)
      fund_train.append(p)
      if p > optimum:
        optimum = p
    optimum_train = []
    for f in fund_train:
      if f == 0:
        optimum_train.append(0)
      else:
        optimum_train.append(optimum)
    funds_train.append(fund_train)
    duration_train.append(start)
    optima_train.append(optimum_train)

  manual_choice = [0.0] * len(funds_data)
  manual_choice[126] = 1.0
  print(choice_loss(funds_train, duration_train, optima_train, manual_choice))

  funds = tf.placeholder(tf.float32)
  duration = tf.placeholder(tf.float32)
  optima = tf.placeholder(tf.float32)
  choice = tf.Variable(choice_init, tf.float32)
  loss = tf.reduce_sum(optima - tf.abs(choice) * funds / tf.reduce_sum(tf.abs(choice))) * duration
  train = tf.train.GradientDescentOptimizer(0.01).minimize(loss)

  assig = {funds: funds_train, optima: optima_train, duration: duration_train}
  sess = tf.Session()
  sess.run(tf.global_variables_initializer())
  for i in range(5000):
    sess.run(train, assig)
    if i % 100 == 0:
      curr_choice, curr_loss = sess.run([choice, loss], assig)
      for j in range(len(curr_choice)):
        curr_choice[j] = abs(curr_choice[j])
      print(i, choice_loss(funds_train, duration_train, optima_train, curr_choice))


if __name__ == '__main__':
  main()
