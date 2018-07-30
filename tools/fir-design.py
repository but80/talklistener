import numpy as np
from scipy import signal, interpolate
import matplotlib.pyplot as plt
import csv
import sys

t0 = np.array([])
f0 = np.array([])
with open('output/test.f0', 'r') as f:
    reader = csv.reader(f, delimiter=':')
    for row in reader:
        t = float(row[0])
        f = float(row[1])
        t0 = np.append(t0, t)
        f0 = np.append(f0, f)

dt0 = t0[1] - t0[0]
oversample = 5
dt0o = dt0 / oversample

f0o = np.array([])
for i in range(len(f0)-1):
    begin = f0[i]
    end = f0[i+1]
    for j in range(oversample):
        f = (begin * (oversample - j) + end * j) / oversample
        f0o = np.append(f0o, f)
f0o = np.append(f0o, f0[len(f0)-1])

base = f0o[0]
f0o = f0o - base

taps = 221
fc = 1.5 # cutoff frequency (change this)
fs = 1. / dt0

lpf = signal.firwin(numtaps=taps, cutoff=fc/(fs/2.), pass_zero=True)
f0f = signal.lfilter(lpf, 1, f0o)
f0f = f0f[taps//2:]

f0o += base
f0f += base

plt.plot(f0o)
plt.plot(f0f)
plt.show()

sys.stdout.write("var firLPF = []float64{")
for i in range(len(lpf)):
    if i % 10 == 0:
        sys.stdout.write("\n\t")
    else:
        sys.stdout.write(" ")
    sys.stdout.write("%+e," % lpf[i])
sys.stdout.write("\n}\n")
