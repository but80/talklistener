#!/usr/bin/perl

my $compIDs = {
    "Miku(V2)"     => "BHHN4EF9BRWTNHAB",
    "Rin_ACT2(V2)" => "BEKF6B63DMXLRECA",
    "Len_ACT2(V2)" => "BMLBDHXXMWYF2MBE",
    "Luka_JPN(V2)" => "BCMDC9MZLKZHZCB4",
    "Luka_ENG(V2)" => "BHLNEE62NRYK3HD2",
    "Iroha(V2)"    => "BMKN7HT9EWTTSMCL",
    "VY1V3"        => "BDRE87E2FTTKTDBA",
    "Yukari"       => "BMGK9EC6G4RPWMB3",
    "IA"           => "BLRGDDR4M3WM2LC6",
    "CUL"          => "BCBG86S4FSYMTCBK",
};
my $vBSs = {
    "Luka_ENG(V2)" => "1",
};

my $s = `cat all-singers.vsqx | grep -B2 -A1 compID`;
$s =~ s/\n//g;
foreach (split("--", $s)) {
    m|<vBS>(\d+)</vBS>.+CDATA\[(.*?)\].+CDATA\[(.*?)\]|;
    $compIDs->{$3} = $2;
    $vBSs->{$3} = $1;
}

print "var singerDefs = map[string]singerDef{\n";
foreach (sort keys %$compIDs) {
    $id = $compIDs->{$_};
    $vBS = $vBSs->{$_} || "0";
    print "\t";
    printf '"%s": {compID:"%s", bs:%s},', $_, $id, $vBS;
    print "\n";
}
print "}\n";
