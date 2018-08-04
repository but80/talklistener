from scipy import signal
import sys

taps = 221
fcs = [ 0.5, 1.0, 1.5, 2.0, 2.5, 3.0 ]
frame_period = 0.005
fs = 1. / frame_period

sys.stdout.write("var FIRLPFCutoffs = []string{\n\t")
for fc in fcs:
    sys.stdout.write("\"%2.1f\"," % fc)
sys.stdout.write("\n}\n")

sys.stdout.write("var firLPF = map[string][]float64{\n")
for fc in fcs:
    lpf = signal.firwin(numtaps=taps, cutoff=fc/(fs/2.), pass_zero=True)
    sys.stdout.write("\t\"%2.1f\": {" % fc)
    for i in range(len(lpf)):
        if i % 10 == 0:
            sys.stdout.write("\n\t\t")
        else:
            sys.stdout.write(" ")
        sys.stdout.write("%+e," % lpf[i])
    sys.stdout.write("\n\t},\n")
sys.stdout.write("}\n")
