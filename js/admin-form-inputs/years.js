var yearsForm = {
    removeYear: function () {
        var year = event.target.parentNode;
        var removedYearChecked = year.querySelector('.year-radio').checked;
        year.remove();
        if (removedYearChecked) {
            var yearsParent = document.getElementById('years');
            var yearRadios = yearsParent.querySelectorAll('.year-radio');
            if (yearRadios.length > 0) {
                yearRadios[yearRadios.length - 1].checked = true;
            }
        }
    },

    addYear: function () {
        var yearsParent = document.getElementById('years');
        var year = document.getElementById('add-year-input').value;
        var yearInputs = yearsParent.querySelectorAll('.year-input');
        for (var i = 0; i < yearInputs.length; i++) {
            var yearI = yearInputs[i].value;
            if (yearI === year) {
                return;
            }
        }
        var newYear = yearsForm.createYear(year, "true");
        newYear.querySelector('.year-radio').focus();
    },

    createYear: function (yearNum, active) {
        var template = document.getElementById("year-template");
        var clone = document.importNode(template.content, true);
        var newYear = clone.querySelector('.form-group');
        newYear.id = 'year-' + yearNum;
        newYear.querySelector('.year-radio').id = 'year-' + yearNum + '-active';
        newYear.querySelector('.year-radio').value = yearNum;
        if (active == "true") {
            newYear.querySelector('.year-radio').checked = true;
        }
        newYear.querySelector('.year-label').htmlFor = 'year-' + yearNum + '-active';
        newYear.querySelector('.year-label').innerText = yearNum;
        newYear.querySelector('.year-input').value = yearNum;
        var years = document.getElementById("year-form-items");
        years.appendChild(clone);
        return newYear;
    },

    initYears: function () {
        var years = document.getElementById("year-form-items").children;
        for (var i = 0; i < years.length; i++) {
            var value = years[i].querySelector(".value").innerText;
            var active = years[i].querySelector(".active").innerText;
            var newYear = yearsForm.createYear(value, active);
            years[i].replaceWith(newYear);
        }
    },

    initAddYearInput: function () {
        document.getElementById('add-year-input').value = new Date().getYear() + 1900;
    },
};

yearsForm.initYears();
yearsForm.initAddYearInput();