#include <limits.h>
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

int main(void) {
  // printf.
  char* message = "Hello, world!";
  printf("%s\n", message);
  printf("Message points to %p.\n", message);

  // Iteration.
  for (int i = 1; i <= 10; ++i) {
    printf("i is %d\n", i);
  }

  // Math.
  double x = 0.34891;
  printf("sin(%f) = %f\n", x, sin(x));

  // Memory allocation.
  char* hostname = malloc(HOST_NAME_MAX);
  gethostname(hostname, HOST_NAME_MAX);
  if (strlen(hostname) > 0) {
    printf("Hostname is %s.\n", hostname);
  } else {
    printf("Hostname is empty.\n");
  }
  free(hostname);

  return 0;
}
