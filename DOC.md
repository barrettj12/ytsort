From https://developers.google.com/youtube/v3/guides/implementation/playlists :

Update a playlist item
----------------------

This example updates a playlist item so that it is the first item in a playlist. This request must be authorized using OAuth 2.0. This example has three steps:

*   **Step 1: Retrieve the appropriate playlist ID**
    
    Call the [`playlist.list`](/youtube/v3/docs/playlists/list) method to retrieve the playlists in the currently authenticated user's channel. The sample request above for [retrieving the current user's playlists](#playlists-retrieve-for-current-user) demonstrates this request. The application calling the API could process the API response to display a list of playlists, using each playlist's ID as a key.
    
*   **Step 2: Retrieve the items in the selected playlist**
    
    Call the `playlistItems.list` method to retrieve the list of videos in the selected playlist. Set the [`playlistId`](/youtube/v3/docs/playlistItems/list#playlistId) parameter's value to the playlist ID that you obtained in step 1.
    
    Each resource in the API response contains an `id` property, which identifies the playlist item ID that uniquely identifies that item. You will use that value to remove an item from the list in the next step.
    
*   **Step 3: Update the selected playlist item**
    
    Call the [`playlistItems.update`](/youtube/v3/docs/playlistItems/update) method to change the video's position in the playlist. Set the [`part`](/youtube/v3/docs/playlistItems/update#part) parameter value to `snippet`. The request body must be a [`playlistItem`](/youtube/v3/docs/playlistItems) resource that at least sets the following values:
    
    *   Set the `id` property to the playlist item ID obtained in step 2.
    *   Set the `snippet.playlistId` property to the playlist ID obtained in step 1.
    *   Set the `snippet.resourceId.kind` property to `youtube#video`.
    *   Set the `snippet.resourceId.videoId` property to the video ID that uniquely identifies the video included in the playlist.
    *   Set the `snippet.position` property to `0` or to whatever position you want the item to appear (using a 0-based index).
    
    The API request below updates a playlist item to be the first item in a playlist. The request body is:
    ```
    {
      "id": "PLAYLIST\_ITEM\_ID",
      "snippet": {
        "playlistId": "PLAYLIST\_ID",
        "resourceId": {
          "kind": "youtube#video",
          "videoId": "VIDEO\_ID"
        },
        "position": 0
      }
    }
    ```
    To complete the request in the APIs Explorer, you need to set values for the `id`, `snippet.playlistId` and `snippet.resourceId.videoId` properties.