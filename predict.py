import cProfile
import tensorflow as tf
import numpy as np

money = 115000

def main():
  funds = []
  with open("get.tsv") as f:
    for line in f:
      funds.append(Fund(line))
  optimum = OptimumFund(funds)
  for f in funds:
    f.createAnnual(optimum.duration)
  optimum.createAnnual(funds)
  # print('PredictStrategy', averageLoss(optimum, funds, PredictStrategy(), money, optimum.duration, 0))
  # print('PastBest', averageLoss(optimum, funds, PastBestStrategy(), money, optimum.duration, 0))
  # print('SmallestLoss', averageLoss(optimum, funds, SmallestLossStrategy(), money, optimum.duration, 0))
  # for i in range(len(funds)):
  #   print('SingleFund(' + str(i) + '):' + str(funds[i].duration), averageLoss(optimum, funds, SingleFundStrategy(i), money, optimum.duration, 0))
  print('PredictStrategy', loss(optimum, funds, PredictStrategy(), money, optimum.duration, 0, optimum.duration // 2))
  # print('PastBest', loss(optimum, funds, PastBestStrategy(), money, optimum.duration, 0, optimum.duration // 2))
  # print('SmallestLoss', loss(optimum, funds, SmallestLossStrategy(), money, optimum.duration, 0, optimum.duration // 2))
  # for i in range(len(funds)):
  #   print('SingleFund(' + str(i) + '):' + str(funds[i].duration), loss(optimum, funds, SingleFundStrategy(i), money, optimum.duration, 0, optimum.duration // 2))


class Fund:
  def __init__(self, line):
    fields = line.strip().split('\t')
    self.name = fields[0]
    self.min = int(fields[1][:-3].replace('.', ''))
    self.raw = []
    for f in fields[4:]:
      self.raw.insert(0, 1 + float(f.replace(',', '.')) / 100)
    self.duration = len(self.raw)
    self.createTable()

  def createTable(self):
    self.table = []
    for start in range(self.duration + 1):
      self.table.append([0] * start)
    for start in range(self.duration, 0, -1):
      self.table[start][start - 1] = self.raw[start - 1]
      for end in range(start - 2, -1, -1):
        self.table[start][end] = self.table[start][end + 1] * self.raw[end]

  def createAnnual(self, max_time):
    self.annual = []
    for start in range(max_time + 1):
      if start > self.duration:
        self.annual.append(self.annual[self.duration])
        continue
      self.annual.append([0] * max_time)
      for end in range(start):
        self.annual[start][end] = self.table[start][end] ** (1.0/((start - end)/12.0))


class OptimumFund:
  def __init__(self, funds):
    self.duration = 0
    for f in funds:
      if f.duration > self.duration:
        self.duration = f.duration

  def createAnnual(self, funds):
    self.annual = [[]] * (self.duration + 1)
    for start in range(self.duration + 1):
      self.annual[start] = [0] * start
      for end in range(start):
        max = 0
        for f in funds:
          if start > f.duration:
            continue
          r = f.annual[start][end]
          if r > max:
            max = r
        self.annual[start][end] = max


def averageLoss(optimum, funds, strategy, money, start, end):
  num = 0
  den = 0
  for time in range(end + 1, start):
    l = loss(optimum, funds, strategy, money, start, end, time)
    if l is None:
      continue
    num += time * l
    den += time
  if den == 0:
    return None
  return num / den * 100.0


def loss(optimum, funds, strategy, money, start, end, time):
  fis = strategy.select(optimum, funds, money, start, time)
  if len(fis) == 0:
    return None
  num = 0
  den = 0
  for fi in fis:
    if time > funds[fi].duration:
      return None
    num += funds[fi].min * funds[fi].table[time][end]
    den += funds[fi].min
  return optimum.annual[time][0] - (num/den) ** (1.0/(time/12.0))


class PastBestStrategy:
  def select(self, optimum, funds, money, start, end):
    fis = list(range(len(funds)))
    fis = sorted(fis, key=lambda fi: -funds[fi].annual[start][end])
    choice = []
    for fi in fis:
      if funds[fi].annual[start][end] == 0:
        break
      if funds[fi].annual[start][end] == 0 or funds[fi].min > money:
        break
      choice.append(fi)
      money -= funds[fi].min
    return choice


class SingleFundStrategy:
  def __init__(self, fund):
    self.fund = fund

  def select(self, optimum, funds, money, start, end):
    return [self.fund]


class SmallestLossStrategy:
  def select(self, optimum, funds, money, start, end):
    loss = []
    for i in range(len(funds)):
      l = averageLoss(optimum, funds, SingleFundStrategy(i), money, start, end)
      if l is None:
        l = 1000000
      loss.append(l)
    fis = list(range(len(funds)))
    fis = sorted(fis, key=lambda fi: loss[fi])
    choice = []
    for fi in fis:
      if funds[fi].annual[start][end] == 0:
        break
      if funds[fi].annual[start][end] == 0 or funds[fi].min > money:
        break
      choice.append(fi)
      money -= funds[fi].min
    return choice

class PredictStrategy:
  def select(self, optimum, funds, money, start, end):
    if start - end <= 1:
      return []
    train_input = []
    train_output = []
    i = 0
    for time in range(end + 1, start):
      for f in funds:
        if f.annual[time][end] == 0:
          continue
        train_output.append(f.annual[time][end])
        input = [0] * (start - end - 1)
        for month in range(time, start):
          if month >= f.duration:
            break
          input[month - time] = f.raw[month]
        train_input.append(input)
        i += 1

    train_input_fn = tf.estimator.inputs.numpy_input_fn(
      {'funds': np.array(train_input)},
      np.array(train_output),
      batch_size=len(train_input),
      num_epochs=None,
      shuffle=True)
    feature_columns = [tf.feature_column.numeric_column('funds', shape=[start - end - 1])]
    estimator = tf.estimator.LinearRegressor(feature_columns=feature_columns) # DNNClassifier(feature_columns=feature_columns, hidden_units=[1024, 512, 256])
    estimator.train(input_fn=train_input_fn, steps=100)

    input = np.zeros((len(funds), start - end - 1))
    for i in range(len(funds)):
      for month in range(time, start):
        if month >= funds[i].duration:
          break
        train_input[i][month - time] = funds[i].raw[month]
    input_fn = tf.estimator.inputs.numpy_input_fn(
      {'funds': input},
      num_epochs=1,
      shuffle=False)
    predictions = estimator.predict(input_fn=input_fn)
    pred = []
    for p in predictions:
      pred.append(p['predictions'][0])
    print(pred)

    fis = list(range(len(funds)))
    fis = sorted(fis, key=lambda fi: -pred[fi])
    choice = []
    for fi in fis:
      if funds[fi].annual[start][end] == 0:
        break
      if funds[fi].annual[start][end] == 0 or funds[fi].min > money:
        break
      choice.append(fi)
      money -= funds[fi].min
    return choice

if __name__ == '__main__':
  main()
  # cProfile.run('main()')
