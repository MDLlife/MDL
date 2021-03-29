sudo npm i -g wine
wine
sudo apt-add-repository 'deb https://dl.winehq.org/wine-builds/ubuntu/ xenial main'
wget -nc https://dl.winehq.org/wine-builds/winehq.key
sudo apt-key add winehq.key\n
sudo apt-add-repository 'deb https://dl.winehq.org/wine-builds/ubuntu/ cosmic main'\n
sudo apt install --install-recommends winehq-stable\n
sudo apt-get install gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64 wine64
