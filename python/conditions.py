from cs50 import get_int

x = get_int("x: ")
y = get_int("y: ")

if x > y:
    print("x is greater than y")
elif x < y:
    print("y is greater than x")
elif x == y:
    print("x is equal to y")
