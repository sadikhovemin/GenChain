from random import randint
import libnum
import sys


class Patient:
	def __init__(self, name, diseaseList, relative):
		self.name = name
		self.diseaseList = diseaseList
		self.relative = relative

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


fatherList = [0,1,0,1,0]
motherList = [0,0,1,0,0]
auntList = [0,0,0,0,0]
grandfatherList = [0,0,1,0,0]

father = Patient("father",fatherList, 1)
mother = Patient("mother",motherList, 1)
aunt = Patient("aunt",auntList, 2)
grandFather = Patient("grandfather",grandfatherList, 2)

familyTree = [father,mother,aunt,grandFather]

sum = 0;
diseaseIndex = 2;
diseasePercentage = 25;


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

if (gcd(g,n*n)!=1):
	print("g is not relatively prime to n*n. Exit...")
	sys.exit()
	

sumCipher = encryption(g,sum,r,n)

relativeLevel = 1


for patient in familyTree:

	print(patient.name , " , Relative Level : " , patient.relative , " , Is it sick : " , patient.diseaseList[diseaseIndex])

	if(patient.relative != relativeLevel):
		relativeLevel+=1
		diseasePercentage//=2


	message1 = diseasePercentage
	message2 = patient.diseaseList[diseaseIndex]

	cipher2 = encryption(g,message2,r,n)

	ciphertotal = HomomorphicMultiplication(cipher2,message1)
	sumCipher = HomomorphicAddition(ciphertotal,sumCipher)
	messTotal = decrption(sumCipher,gLambda,n,gMu)

	print("================")
	print("Currently sick probability : %" , messTotal)
	print("================")

