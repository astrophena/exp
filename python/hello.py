# https://cs50.harvard.edu/x/2021/weeks/6/

from cs50 import get_string, get_int

answer = get_string("What's your name? ")
print(f"Hello, {answer}!")

x = get_int("x: ")
print(f"x = {x}")
y = get_int("y: ")
print(f"y = {y}")

if x > y:
    print("x > y")
elif x < y:
    print("x < y")
