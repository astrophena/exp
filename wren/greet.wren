class Greeter {
  construct new(name) {
    _name = name
  }

  greet() {
    System.print("Hello, %(_name)!")
  }
}

var greeter = Greeter.new("Ilya")
greeter.greet()
