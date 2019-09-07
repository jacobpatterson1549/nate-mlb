var playerSearch = {
    showModal: function (show) {
        var addPlayerModal = document.getElementById('addPlayerModal');
        addPlayerModal.classList.toggle('d-none', !show);
        var openModalButton = document.getElementById('openPlayerSearchModal')
        openModalButton.classList.toggle('d-none', show);
        // addPlayerModal.style.toggle('display: block; padding-right: 15px;', show);
        // addPlayerModal.setAttribute("aria-modal", show);
        if (show) {
            playerSearch.initActivePlayersCB();
            // var modalBackdrop = document.createElement('div')
            // modalBackdrop.classList.add('modal-backdrop');
            // document.body.appendChild(modalBackdrop);
        } else {
            // var modalBackdrop = document.getElementById('modalBackdrop show');
            // modalBackdrop.remove();
        }
    },

    initActivePlayersCB: function() {
        var playerType = document.getElementById('select-player-type').value;
        var isMlbPlayerType = [2, 3].includes(parseInt(playerType)); // PlayerTypeHitter, PlayerTypePitcher
        var activePlayersOnlyGroup = document.getElementById("apo-group");
        activePlayersOnlyGroup.classList.toggle('d-none', !isMlbPlayerType);
    },

    add: function () {
        var playerSearchResults = document.getElementById('player-search-results');
        playerSearchResults = playerSearchResults.getElementsByClassName('form-check');
        for (var i = 0; i < playerSearchResults.length; i++) {
            var psr = playerSearchResults[i];
            if (psr.querySelector('.psr-radio').checked) {
                var sourceID = psr.querySelector('.psr-source-id').value;
                var playerName = psr.querySelector('.psr-player-name').value;
                var newPlayer = playersForm.add(playerName, sourceID);
                newPlayer.focus();
                playerSearch.showModal(false);
                return;
            }
        }
    },

    clear: function () {
        var resultsDiv = document.getElementById('player-search-results-output');
        resultsDiv.innerHTML = '';
    },

    submit: function (event) {
        event.preventDefault();
        playerSearch.setNewPlayerSelected(false);
        var playerType = document.getElementById('select-player-type').value;
        var formData = new FormData(event.target);
        formData.append('pt', playerType);
        var url = window.location.pathname + "/search?" + new URLSearchParams(formData);
        fetch(url, {
            method: 'GET',
        }).then(async res => {
            if (res.status == 200) {
                return res.json();
            } else {
                var message = await res.text();
                return Promise.reject(message);
            }
        }).then(playerSearchResults => {
            if (playerSearchResults != null && playerSearchResults.length > 0) {
                playerSearch.success(playerSearchResults);
                return Promise.resolve();
            } else {
                return Promise.reject('No results');
            }
        }).catch(err => {
            var resultsDiv = document.getElementById('player-search-results-output');
            resultsDiv.innerHTML = err;
        });
    },

    success: function (playerSearchResults) {
        var template = document.getElementById("player-search-results-template");
        var playerSearchResultsDiv = document.importNode(template.content, true);
        var playerSearchResultsFieldSet = playerSearchResultsDiv.getElementById('player-search-results');
        for (var i = 0; i < playerSearchResults.length; i++) {
            var playerSearchResult = playerSearchResults[i];
            var template2 = document.getElementById("player-search-result-template");
            var playerSearchResultDiv = document.importNode(template2.content, true);
            var psr = playerSearchResultDiv.querySelector('.form-check');
            psr.querySelector('.psr-radio').id = 'psr-' + playerSearchResult.SourceID;
            psr.querySelector('.psr-label').htmlFor = 'psr-' + playerSearchResult.SourceID;
            psr.querySelector('.psr-label-name').innerText = playerSearchResult.Name;
            psr.querySelector('.psr-label-details').innerText = playerSearchResult.Details;
            psr.querySelector('.psr-source-id').value = playerSearchResult.SourceID;
            psr.querySelector('.psr-player-name').value = playerSearchResult.Name;
            playerSearchResultsFieldSet.appendChild(psr);
            if (i == 0 && playerSearchResults.length == 1) {
                psr.querySelector('.psr-radio').checked = true;
            }
        }
        var resultsDiv = document.getElementById('player-search-results-output');
        resultsDiv.innerHTML = '';
        resultsDiv.appendChild(playerSearchResultsFieldSet);
        playerSearch.setNewPlayerSelected(playerSearchResults.length == 1);
    },

    setNewPlayerSelected: function (isSelected) {
        var addPlayerButton = document.getElementById('modal-add-player-button');
        addPlayerButton.disabled = !isSelected;
    },
};
