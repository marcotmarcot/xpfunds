import random
import sys

def main():
  funds = []
  with open("get.tsv") as f:
    for line in f:
      funds.append(Fund(line))
  optimum = OptimumFund(funds)
  for f in funds:
    f.createAnnual(optimum.duration)
  optimum.createAnnual(funds)
  strategies = [RandomStrategy(), BestStrategy(), WorstStrategy()]
  for s in strategies:
    for num_funds in range(1, 20):
      for min_time in range(12):
        l, d = averageLossAndDiff(optimum, funds, s, num_funds, min_time, optimum.duration, 0)
        print(s.name, '\t',  num_funds, '\t', min_time, '\t', l, '\t', d)
        sys.stdout.flush()
  for i in range(len(funds)):
    s = ConstStrategy([i])
    l, d = averageLossAndDiff(optimum, funds, s, 0, 0, optimum.duration, 0)
    print(s.name, '\t0\t0\t', l, '\t', d)

class Fund:
  def __init__(self, line):
    fields = line.strip().split('\t')
    self.name = fields[0]
    self.min = int(fields[1][:-3].replace('.', ''))
    self.raw = []
    for f in fields[4:]:
      self.raw.append(1 + float(f.replace(',', '.')) / 100)
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


def averageLossAndDiff(optimum, funds, strategy, num_funds, min_time, start, end):
  ln = 0
  ld = 0
  dn = 0
  dd = 0
  for time in range(end + 1, start):
    duration = time - end
    l, d = lossAndDiff(optimum, funds, strategy, num_funds, min_time, start, end, time)
    if l is not None:
      ln += l * duration
      ld += duration
    if d is not None:
      dn += d * duration
      dd += duration
  if ld == 0:
    l = None
  else:
    l = ln / ld
  if dd == 0:
    d = None
  else:
    d = dn / dd
  return l, d


def lossAndDiff(optimum, funds, strategy, num_funds, min_time, start, end, time):
  fis = strategy.select(optimum, funds, num_funds, min_time, start, time)
  if len(fis) == 0:
    return None, None
  numl = 0
  for fi in fis:
    if time > funds[fi].duration:
      return None, None
    numl += funds[fi].annual[time][end]
  l = optimum.annual[time][end] - numl/len(fis)
  numd = 0
  for fi in fis:
    if time + 1 > funds[fi].duration:
      return l, None
    numd += funds[fi].annual[start][time]
  d = numd/len(fis) - numl/len(fis)
  return l, d


class BestStrategy:
  def __init__(self):
    self.name = 'Best'

  def select(self, optimum, funds, num_funds, min_time, start, end):
    return sortAndPick(funds, num_funds, min_time, start, end, lambda fi: -funds[fi].annual[start][end])


class WorstStrategy:
  def __init__(self):
    self.name = 'Worst'

  def select(self, optimum, funds, num_funds, min_time, start, end):
    annual = []
    for fi in range(len(funds)):
      if funds[fi].annual[start][end] == 0:
        annual.append(10000)
      else:
        annual.append(funds[fi].annual[start][end])
    return sortAndPick(funds, num_funds, min_time, start, end, lambda fi: annual[fi])


class ConstStrategy:
  def __init__(self, funds):
    self.funds = funds
    self.name = 'Const' + str(funds).replace(' ', '')

  def select(self, optimum, funds, num_funds, min_time, start, end):
    return self.funds


class RandomStrategy:
  def __init__(self):
    self.name = 'Random'

  def select(self, optimum, funds, num_funds, min_time, start, end):
    return sortAndPick(funds, num_funds, min_time, start, end, lambda fi: random.uniform(0, 1))


def sortAndPick(funds, num_funds, min_time, start, end, sortFn):
  fis = sorted(list(range(len(funds))), key=sortFn)
  choice = []
  for fi in fis:
    if end + min_time > funds[fi].duration:
      continue
    choice.append(fi)
    if len(choice) >= num_funds:
      break
  return choice


if __name__ == '__main__':
  main()
