function movePlayerUp() {
    var player1 = event.target.parentNode;
    var player2 = event.target.parentNode.previousElementSibling;
    if (player2 != null) {
        var players = player1.parentNode;
        players.insertBefore(player1, player2);
        changePlayerDisplayOrder(player1.id, -1);
        changePlayerDisplayOrder(player2.id, +1);
    }
}

function movePlayerDown() {
    var player1 = event.target.parentNode;
    var player2 = event.target.parentNode.nextElementSibling;
    if (player2 != null) {
        var players = player1.parentNode;
        players.insertBefore(player2, player1);
        changePlayerDisplayOrder(player2.id, -1);
        changePlayerDisplayOrder(player1.id, +1);
    }
}

function removePlayer() {
    var removePlayer = event.target.parentNode;
    for (var player = document.getElementById(removePlayer.id).nextElementSibling; player != null; player = document
        .getElementById(player.id).nextElementSibling) {
        changePlayerDisplayOrder(player.id, -1);
    }
    removePlayer.remove();
}

function addPlayer(playerName, sourceID) {
    var playerType = document.getElementById('select-player-type').value;
    var friendID = document.getElementById('select-friend').value;
    var maxID = 0;
    var maxDisplayOrder = 0;
    var players = document.getElementById('players');
    var playerElements = players.getElementsByClassName('form-group');
    for (var i = 0; i < playerElements.length; i++) {
        var player = playerElements[i];
        var ID = parseInt(player.querySelector('.player-id').value);
        if (ID > maxID) {
            maxID = ID;
        }
        if (playerType === player.querySelector('.player-player-type').value &&
            friendID === player.querySelector('.player-friend-id').value) {
            var playerDisplayOrder = parseInt(player.querySelector('.player-display-order').value);
            if (playerDisplayOrder > maxDisplayOrder) {
                maxDisplayOrder = playerDisplayOrder;
            }
        }
    }
    return createPlayer(maxID + 1, playerName, sourceID, maxDisplayOrder + 1, playerType, friendID);
}

function createPlayer(id, playerName, sourceID, displayOrder, playerType, friendID) {
    var template = document.getElementById("player-template");
    var clone = document.importNode(template.content, true);
    var player = clone.querySelector('.form-group');
    player.id = 'player-' + id;
    player.querySelector('.player-name-label').innerText = playerName;
    player.querySelector('.player-source-id').name = 'player-' + id + '-source-id';
    player.querySelector('.player-source-id').value = sourceID;
    player.querySelector('.player-display-order').name = 'player-' + id + '-display-order';
    player.querySelector('.player-display-order').value = displayOrder;
    player.querySelector('.player-player-type').name = 'player-' + id + '-player-type';
    player.querySelector('.player-player-type').value = playerType;
    player.querySelector('.player-friend-id').name = 'player-' + id + '-friend-id';
    player.querySelector('.player-friend-id').value = friendID;
    player.querySelector('.player-id').value = id;
    var scoreCategories = document.getElementById("players");
    var scoreCategory = scoreCategories.querySelector('.player-type-' + playerType);
    var friendScore = scoreCategory.querySelector('.friend-id-' + friendID);
    friendScore.appendChild(clone);
    return player;
}

function changePlayerDisplayOrder(playerID, delta) {
    var player = document.getElementById(playerID);
    var playerDisplayOrderElement = player.querySelector('.player-display-order');
    var playerDisplayOrder = playerDisplayOrderElement.value;
    playerDisplayOrderElement.value = parseInt(playerDisplayOrder) + delta;
}

function refreshVisiblePlayers() {
    var playerType = document.getElementById('select-player-type').value;
    var friendID = document.getElementById('select-friend').value;
    var players = document.getElementById('players');
    var playerElements = players.getElementsByClassName('form-group');
    for (var i = 0; i < playerElements.length; i++) {
        var player = playerElements[i];
        var currentPlayerType = player.querySelector('.player-player-type').value;
        var currentFriendID = player.querySelector('.player-friend-id').value;
        if (playerType === currentPlayerType && friendID === currentFriendID) {
            player.classList.remove('d-none');
        } else {
            player.classList.add('d-none');
        }
    }
    clearPlayerSearch();
}

function initPlayers() {
    var playerTypes = document.getElementById("players").children;
    for (var i = 0; i < playerTypes.length; i++) {
        var friendScores = playerTypes[i].children;
        for (var j = 0; j < friendScores.length; j++) {
            var playerScores = friendScores[j].children;
            for (var k = 0; k < playerScores.length; k++) {
                var id = playerScores[k].querySelector(".id").innerText;
                var playerName = playerScores[k].querySelector(".playerName").innerText;
                var sourceID = playerScores[k].querySelector(".sourceID").innerText;
                var displayOrder = playerScores[k].querySelector(".displayOrder").innerText;
                var playerType = playerScores[k].querySelector(".playerType").innerText;
                var friendID = playerScores[k].querySelector(".friendID").innerText;
                var newPlayer = createPlayer(id, playerName, sourceID, displayOrder, playerType, friendID);
                playerScores[k].replaceWith(newPlayer);
            }
        }
    }
}

initPlayers();
refreshVisiblePlayers(); // initialize visible player scores to first friendScore.