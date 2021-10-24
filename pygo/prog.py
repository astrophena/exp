import ctypes

lib = ctypes.cdll.LoadLibrary("./lib.so")
print(f"Add: {lib.Add(1, 2)}")

lib.Version.restype = ctypes.c_char_p
print(f"Version: {lib.Version()}")
