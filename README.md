# mac-spoofer

A basic mac spoofer for windows.

### How it works

It modifies the `Computer\HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}\ADAPTER_ID` registry key (where `ADAPTER_ID` is the # of whatever adapter you want to spoof). It sets the "NetworkAddress" string value to be whatever wanted.

### Issues

Spoofing the mac address will come with issues. To summarize, the first octet needs to be X2, X6, XA or XE, because the first 2 bits (LAA and multicast) have to be a certain value for spoofed addresses. The LAA (Locally Administered Address) Bit needs to be 1, and the uni/multicast bit needs to be 0 (unicast).

This can be bypassed by more complicated trickery, but falls out of the range of a basic mac spoofer.