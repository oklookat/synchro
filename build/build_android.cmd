mkdir ..\dist
cd ../commander
gomobile bind -target android -androidapi 24 -javapkg "com.oklookat.synchro" -o ../dist/core.aar
