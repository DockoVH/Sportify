const lozinkaInput = document.querySelector('#signup-input-lozinka')
const prikaziLozinkuCheckbox = document.querySelector('#signup-prikazi-lozinku-checkbox')
const ponoviLozinkuInput = document.querySelector('#signup-input-ponovi-lozinku')
const prikaziPonoviLozinkuCheckbox = document.querySelector('#signup-prikazi-ponovi-lozinku-checkbox')
const usernameInput = document.querySelector('#signup-input-username')

lozinkaInput.type = prikaziLozinkuCheckbox.checked ? "text" : "password"
prikaziLozinkuCheckbox.addEventListener('change', () => {
    lozinkaInput.type = prikaziLozinkuCheckbox.checked ? "text" : "password"
})

prikaziPonoviLozinkuCheckbox.addEventListener('change', () => {
    ponoviLozinkuInput.type = prikaziPonoviLozinkuCheckbox.checked ? "text" : "password"
})

usernameInput.oninput =  () => {
    usernameInput.value = usernameInput.value.toLowerCase().replace(/[^a-z0-9_]/g, '')
}