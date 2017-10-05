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
  strategies = [
    PastBestStrategy(),
    SmallestLossStrategy()]
  for i in range(len(funds)):
    strategies.append(ConstStrategy([i]))
  for inlen in range(5):
    for outlen in range(5):
      strategies.append(PredictStrategy(inlen * 24 + 1, outlen * 24 + 1))
  # strategies.append(PredictStrategy(74, 13))
  results = []
  for s in strategies:
    v = averageLoss(optimum, funds, s, money, optimum.duration, 0)
    # v = loss(optimum, funds, s, money, optimum.duration, 0, 3 * optimum.duration // 4)
    print(s.name, v)


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
  return num / den


def loss(optimum, funds, strategy, money, start, end, time):
  fis = strategy.select(optimum, funds, money, start, time)
  if len(fis) == 0:
    return None
  num = 0
  den = 0
  for fi in fis:
    if time > funds[fi].duration:
      return None
    num += funds[fi].table[time][end]
  return optimum.annual[time][0] - (num/len(fis)) ** (1.0/(time/12.0))


class PastBestStrategy:
  def __init__(self):
    self.name = 'Best'

  def select(self, optimum, funds, money, start, end):
    return sortAndPick(funds, money, start, end, lambda fi: -funds[fi].annual[start][end])


class ConstStrategy:
  def __init__(self, funds):
    self.funds = funds
    self.name = 'Const' + str(funds).replace(' ', '')

  def select(self, optimum, funds, money, start, end):
    return self.funds


class SmallestLossStrategy:
  def __init__(self):
    self.name = 'Loss'

  def select(self, optimum, funds, money, start, end):
    loss = []
    for i in range(len(funds)):
      l = averageLoss(optimum, funds, ConstStrategy([i]), money, start, end)
      if l is None:
        l = 1000000
      loss.append(l)
    return sortAndPick(funds, money, start, end, lambda fi: loss[fi])


class PredictStrategy:
  def __init__(self, inlen, outlen):
    self.inlen = inlen
    self.outlen = outlen
    self.name = 'Predict(' + str(inlen) + ',' + str(outlen) + ')'

  def select(self, optimum, funds, money, start, end):
    if start - end < self.inlen + self.outlen:
      return []
    train_input = []
    train_output = []
    for time in range(end + self.outlen, start - self.inlen):
      for f in funds:
        if time + self.inlen > f.duration:
          continue
        train_output.append(f.annual[time][time - self.outlen])
        input = []
        for month in range(time, time + self.inlen):
          input.append(f.raw[month])
        train_input.append(input)

    if len(train_input) == 0:
      return []
    train_input_fn = tf.estimator.inputs.numpy_input_fn(
      {'funds': np.array(train_input)},
      np.array(train_output),
      batch_size=len(train_input),
      num_epochs=None,
      shuffle=True)
    feature_columns = [tf.feature_column.numeric_column('funds', shape=[self.inlen])]
    estimator = tf.estimator.LinearRegressor(feature_columns=feature_columns)
    estimator.train(input_fn=train_input_fn, steps=1000)

    pred_input = []
    input_fi = []
    for i in range(len(funds)):
      if end + self.inlen > funds[i].duration:
        continue
      input = []
      for month in range(end, end + self.inlen):
        input.append(funds[i].raw[month])
      pred_input.append(input)
      input_fi.append(i)
    input_fn = tf.estimator.inputs.numpy_input_fn(
      {'funds': np.array(pred_input)},
      num_epochs=1,
      shuffle=False)
    predictions = estimator.predict(input_fn=input_fn)
    pred = [0] * len(funds)
    i = 0
    for p in predictions:
      pred[input_fi[i]] = p['predictions'][0]
      i += 1

    choice = sortAndPick(funds, money, start, end, lambda fi: -pred[fi])
    return choice


def sortAndPick(funds, money, start, end, sortFn):
  fis = sorted(list(range(len(funds))), key=sortFn)
  choice = []
  max = 0
  for fi in fis:
    if funds[fi].annual[start][end] == 0:
      break
    if funds[fi].min > max:
      max = funds[fi].min
    if (len(choice) + 1) * max > money:
      break
    choice.append(fi)
  return choice


if __name__ == '__main__':
  main()
  # cProfile.run('main()')
