import re
import subprocess

# Find all the so files defined
sore = re.compile(".+warning: ([a-z0-9\.]+), needed.+")
libs = []
with open("buildlog") as f:
    line = f.readline()
    while line != "":
        match = sore.match(line)
        if match != None:
            lib = match.groups()[0]
            libs.append(lib)
        line = f.readline()

if len(libs) == 0:
    raise Exception()

aptlist = [] 
findlist = []
for lib in libs:
    libName = lib.split('.')[0]
    if libName == "libdc1394":
        libName += '-22'
    aptlist.append(libName+'-dev')
    findlist.append('-name ' + lib.strip())
print('sudo apt install -y ' + ' '.join(aptlist))
collect_required_deps = ['find', '/lib', '/usr/lib']
collect_postfix = ['-exec', 'cp', '\'{}\'', '/tmp/export_deps/', '\;']

def ssh_call(args):
    final_args = ["ssh", "study"] + args
    print("Going to call: " + str(final_args))
    return subprocess.call(final_args)

def archive():
    ssh_call(["rm", '-rf', '/tmp/export_deps'])
    ssh_call(["mkdir", "/tmp/export_deps"])
    for item in findlist:
        ssh_call(collect_required_deps + [item] + collect_postfix)
    ssh_call(['tar', 'zcvf', '/tmp/exported_deps.tar.gz', '/tmp/export_deps/*'])

archive()
subprocess.call(['scp', 'study:/tmp/exported_deps.tar.gz', '.'])
subprocess.call(['tar', 'zxvf', 'exported_deps.tar.gz', '-C', './rpi_libs/sharelib'])

subprocess.call(['ls'])
def rename():
    for item in libs:
        subprocess.call(['mv', './rpi_libs/sharelib/'+item, './rpi_libs/sharelib/'+'.'.join(item.split('.')[:2])])

#rename()
s = ''
for i in range(0, len(libs)):
    if i != 0:
        s += ' '
    s += '-l'+libs[i].split('.')[0]
print(s)
