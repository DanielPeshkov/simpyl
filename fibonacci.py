def fibonacci(x):
    if x == 0:
        return 0
    
    if x == 1:
        return 1
    
    return fibonacci(x - 1) + fibonacci(x - 2)

result = fibonacci(36)
print(result)