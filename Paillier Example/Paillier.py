from random import randint
import libnum
import sys

def gcd(a,b):
    """Compute the greatest common divisor of a and b"""
    while b > 0:
        a, b = b, a % b
    return a
    
def lcm(a, b):
    """Compute the lowest common multiple of a and b"""
    return a * b // gcd(a, b)

def L(x,n):
	return ((x-1)//n)

def encryption(g,message,r,n):
	k1 = pow(g, message, n*n)
	k2 = pow(r, n, n*n)

	return (k1 * k2) % (n*n)

def decrption(cipher,gLambda,n,gMu):
	l = (pow(cipher, gLambda, n*n)-1) // n
	return (l * gMu) % n

def HomomorphicMultiplication(cipher1,message2):
	return (pow(cipher1,message2)) % (n*n)

def HomomorphicAddition(cipher1,cipher2):
	return (cipher1 * cipher2) % (n*n)
	

"""Pick p and q"""

p=17
q=19


if (p==q):
	print("P and Q cannot be the same")
	sys.exit()

"""Calculation for Paillier Formulas"""

n = p*q
gLambda = lcm(p-1,q-1)
g = randint(20,150)
r = randint(20,150)
l = (pow(g, gLambda, n*n)-1)//n
gMu = libnum.invmod(l, n)

if (gcd(g,n*n)==1):
	print("g is relatively prime to n*n")
else:
	print("g is not relatively prime to n*n. Exit...")
	sys.exit()


message1 = 10
message2 = 2



cipher1 = encryption(g,message1,r,n)
mess = decrption(cipher1,gLambda,n,gMu)

cipher2 = encryption(g,message2,r,n)
mess2= decrption(cipher2,gLambda,n,gMu)

ciphertotal = HomomorphicMultiplication(cipher1,message2)
messTotal = decrption(ciphertotal,gLambda,n,gMu)

print("p=",p,"\tq=",q)
print("g=",g,"\tr=",r)
print("================")
print("Mu:\t\t",gMu,"\tgLambda:\t",gLambda)
print("================")
print("Public key (n,g):\t\t",n,g)
print("Private key (lambda,mu):\t",gLambda,gMu)
print("================")
print("Message 1 :\t",message1)
print("Cipher:\t\t",cipher1)
print("Decrypted:\t",mess)
print("================")
print("Message 2 :\t",message2)
print("Cipher:\t\t",cipher2)
print("Decrypted:\t",mess2)
print("================")
print("Sum Of Two Messages :\t\t",message1 * message2)
print("Cipher:\t\t",ciphertotal)
print("Decrypted:\t",messTotal)
print("================")