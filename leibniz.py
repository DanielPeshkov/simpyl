def main(terms):
    pi = 0.
    sign = 1.
    for n in range(terms):
        pi = pi + sign/(n*2. + 1)
        sign = -sign

    return 4 * pi

terms = 100000000
result = main(terms)
print(result)