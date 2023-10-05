# Kubernetes Firewall Operand

# Purpose

The purpose of this repo is to provide an operand to control firewalls from Kubernetes clusters. The current use case is to open port forwards in a Juniper SRX340 by control of Services. Every Service using NodePort, and Special Annotations should have a Port Forwarding Rule created on the SRX. This is Alpha Software.
