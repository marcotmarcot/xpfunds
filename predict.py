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
import cProfile
# import tensorflow as tf

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
  print(optimum.annual[optimum.duration][0])
  print(optimum.annual[optimum.duration - 1][0])
  print('python')
  print('PastBest', averageLoss(optimum, funds, PastBestStrategy(), money))
  # for i in range(len(funds)):
  #   print(i, averageLoss(optimum, funds, SingleFundStrategy(i), money))


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


def averageLoss(optimum, funds, strategy, money):
  num = 0
  den = 0
  for time in range(1, optimum.duration):
    l = loss(optimum, funds, strategy, money, time)
    if l is None:
      continue
    num += time * l
    den += time
  if den == 0:
    return None
  return num / den * 100.0


def loss(optimum, funds, strategy, money, time):
  fis = strategy.select(optimum.duration, funds, money, time)
  num = 0
  den = 0
  for fi in fis:
    if time > funds[fi].duration:
      return None
    print(funds[fi].min, funds[fi].table[time][0])
    num += funds[fi].min * funds[fi].table[time][0]
    den += funds[fi].min
  print(optimum.annual[time][0])
  print((num/den) ** (1.0/(time/12.0)))
  print(optimum.annual[time][0] - (num/den) ** (1.0/(time/12.0)))
  return optimum.annual[time][0] - (num/den) ** (1.0/(time/12.0))


class PastBestStrategy:
  def select(self, max_time, funds, money, time):
    fis = list(range(len(funds)))
    fis = sorted(fis, key=lambda fi: -funds[fi].annual[max_time][time])
    choice = []
    for fi in fis:
      if funds[fi].annual[max_time][time] == 0 or funds[fi].min > money:
        break
      choice.append(fi)
      money -= funds[fi].min
      print(fi, funds[fi].min, funds[fi].annual[max_time][time])
    return choice


class SingleFundStrategy:
  def __init__(self, fund):
    self.fund = fund

  def select(self, max_time, funds, money, time):
    return [self.fund]


if __name__ == '__main__':
  main()
  # cProfile.run('main()')
