# synchro

Music streaming services utilities.

You can transfer music between services, delete liked tracks, and so on.

I did this project primarily for myself. I wanted to make different UIs, applications, and so on. The ultimate goal was to create something like a unified streaming account, so that no matter what streaming service you were on, everything would be easily transferred and synchronized.

Despite the ambition, I've now settled on CLI, and simpler things, this hell development is getting tiresome.

The main problem is the API of streaming services. There are only two official APIs (Spotify, Deezer), but they work terribly. Spotify often gives errors like 403 despite VPN, and Deezer API is just disabled for everyone, for some reason. The rest of the APIs are made artisanally through tracking HTTP requests.

If you decide to make similar project, please pay attention to two packages: [linker](./linking/linker/) and [linkerimpl](./linking/linkerimpl/). These packages match entities on different streaming services, and do it better than any other services (like TuneMyMusic, Soundiiz, and so on), in my opinion. I recommend you to take note.

This project is either frozen or will be updated, occasionally.
