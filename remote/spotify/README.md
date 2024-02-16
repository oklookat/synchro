# Playlist visibility sync

1. **When creating or updating a playlist through the Spotify Web API, setting the attribute “public” to true will publish the playlist on the user’s profile, which means it will appear under “public playlists”. This will also make the playlist visible in search results.**

2. **Note that the public attribute does not refer to access control, modifying access is currently not possible through the WebAPI, so anyone with the link to the playlist can access it unless it’s made private through for instance the desktop client.**

https://developer.spotify.com/documentation/web-api/concepts/playlists
