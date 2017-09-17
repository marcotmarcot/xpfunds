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
import numpy as np

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
    if start <= end or start > len(self.prof):
      return 0
    return self.table[start][end]

  def prof_list(self, end, max_time):
    prof = self.prof[end:]
    return prof + [0.0] * (max_time - len(prof))

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
  with open('get.tsv') as f:
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

  funds_list_train = []
  one_hots_train = []
  for end in range(1, max_time):
    fund_train = []
    optimum = 0
    oi = 0
    for i in range(len(funds_data)):
      p = funds_data[i].profitability(end, 0)
      # fund_train.append(funds_data[i].prof_list(end, max_time))
      fund_train.append(funds_data[i].profitability(funds_data[i].time(), end))
      if p > optimum:
        optimum = p
        oi = i
    optimum_train = []
    for f in fund_train:
      if f == 0:
        optimum_train.append(0)
      else:
        optimum_train.append(optimum)
    one_hot_train = [0.0] * len(funds_data)
    one_hot_train[oi] = 1.0
    funds_list_train.append(fund_train)
    one_hots_train.append(one_hot_train)

  manual_choice = [0.0] * len(funds_data)
  manual_choice[126] = 1.0
  print(choice_loss(funds_train, duration_train, optima_train, manual_choice))

  funds_list_train
  one_hots_train
  feature_columns = [tf.feature_column.numeric_column('funds', shape=[len(funds_data)])]
  estimator = tf.estimator.LinearRegressor(feature_columns=feature_columns, label_dimension=len(funds_data))
  input_fn = tf.estimator.inputs.numpy_input_fn(
    {'funds': funds_list_train}, one_hots_train, batch_size=(max_time - 1), num_epochs=None, shuffle=True)
  train_input_fn = tf.estimator.inputs.numpy_input_fn(
    {'funds': funds_list_train}, one_hots_train, batch_size=(max_time - 1), num_epochs=1000, shuffle=False)

  estimator.train(input_fn=input_fn, steps=10)

  print(estimator.evaluate(input_fn=train_input_fn))

  predict_input_fn = tf.estimator.inputs.numpy_input_fn({'funds': np.array([fund_train])}, num_epochs=1, shuffle=False)
  predictions = estimator.predict(input_fn=predict_input_fn)
  for i in predictions:
    print(i)


if __name__ == '__main__':
  main()
