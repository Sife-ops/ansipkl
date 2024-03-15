#!/usr/bin/env python3

import ast
import sys

# "./ansible/lib/ansible/modules/setup.py"

try:
    p = sys.argv[1]

    f = open(p)

    r = ast.parse(f.read(), p)

    hasDoc = False
    for node0 in r.body:
        if hasattr(node0, "targets"):
            for node1 in node0.targets:
                if hasattr(node1, "id") and node1.id == "DOCUMENTATION":
                    hasDoc = True
                    print(node0.value.value)

    if not hasDoc:
        print("TODO_NODOC")

except:
    print("TODO_EXCEPTION")
