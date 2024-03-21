#!/usr/bin/env python3

import ast
import sys
from pprint import pprint

try:
    p = sys.argv[1]
    # p = "./ansible/lib/ansible/playbook/base.py"

    c = sys.argv[2]
    # c = "Base"

    f = open(p)

    r = ast.parse(f.read(), p)

    for s in r.body:
        if type(s) != ast.ClassDef:
            continue
        if s.name != c:
            continue

        for s in s.body:
            if type(s) != ast.Assign:
                continue
            if len(s.targets) < 1:
                continue
            if not hasattr(s.value, "keywords"):
                continue

            # print("===")

            if len(s.targets) != 1:
                continue
            f = s.targets[0].id

            # print("field:")
            # print(f)

            isa = None
            # print("keywords:")
            for k in s.value.keywords:
                # pprint(vars(k))
                if k.arg == "isa":
                    isa = k.value.value
            if isa is None:
                continue

            # print("isa:")
            # print(isa)

            ty = "String"
            if isa == "int":
                ty = "Int"
            elif isa == "bool":
                ty = "Boolean"
            elif isa == "list" or isa == "dict":
                ty = "Dynamic"

            # print("formatted:")
            # pprint(vars(s))
            # todo get lineno as doc
            print("{}: {}?".format(f, ty))

except:
    print("TODO_EXCEPTION")

