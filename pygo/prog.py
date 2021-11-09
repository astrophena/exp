import ctypes

lib = ctypes.cdll.LoadLibrary("./lib.so")
print(f"Add: {lib.Add(1, 2)}")

lib.Version.restype = ctypes.c_char_p
print(f"Version: {lib.Version()}")

lib.WatchTime.restype = ctypes.c_char_p
videoID = "h3h035Eyz5A"  # Sia - Unstoppable (Lyrics)
print(f"Watch time: {lib.WatchTime(videoID)}")
