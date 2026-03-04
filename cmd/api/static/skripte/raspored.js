const navbarProfil = document.querySelector('#navbar-profil')
const navbarNapraviOglas = document.querySelector('#napravi-oglas')
const subbarProfil = document.querySelector('#subbar-profil')
const subbarNapraviOglas = document.querySelector('#subbar-napravi-oglas')

if (navbarProfil)
{
    navbarProfil.addEventListener('mouseenter', (e) => {
        subbarProfil.classList.toggle('translate-y-full', true)
    })
    
    navbarProfil.addEventListener('mouseleave', (e) => {
        subbarProfil.classList.toggle('translate-y-full', false)
    })
}

if (navbarNapraviOglas)
{
    navbarNapraviOglas.addEventListener('mouseenter', (e) => {
        subbarNapraviOglas.classList.toggle('translate-y-full', true)
    })
    
    navbarNapraviOglas.addEventListener('mouseleave', (e) => {
        subbarNapraviOglas.classList.toggle('translate-y-full', false)
    })
}