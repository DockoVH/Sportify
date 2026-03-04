package server

import (
	"net/http"
	"strings"
	"log"
	"fmt"
	"time"
	"slices"
	"errors"
	"strconv"
	"math"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"Sportify/internal/db"
	"Sportify/internal/views"
)

const (
	jwtKljuc = "jwtkljuc"
)

var (
	lozinkaSpecKarakteri = []rune { '!', ' ', '@', '#', '$', '%', '&' }
)

type Server struct {
	port int
}

type KorisnikClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (server *Server) HandlerInit(mongoClient *mongo.Client) http.Handler {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")
		fs.ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleHome(mongoClient, w, r)
	})

	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		views.Signup("").Render(r.Context(), w)
	})

	mux.HandleFunc("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		handleRegistrujKorisnika(mongoClient, w, r)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		views.Login("").Render(r.Context(), w)
	})

	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		handlePrijaviKorisnika(mongoClient, w, r)
	})

	mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		handleOdjaviKorisnika(mongoClient, w, r)
	})

	mux.HandleFunc("/oglasi/", func(w http.ResponseWriter, r *http.Request) {
		handleSviOglasi(mongoClient, w, r)
	})

	mux.HandleFunc("/api/oglasiFilter", func(w http.ResponseWriter, r *http.Request) {
		handleOglasiFilter(mongoClient, w, r)
	})

	mux.HandleFunc("/oglas/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlePrikazOglasa(mongoClient, w, r)
	})

	mux.HandleFunc("/api/dodajKomentar/{oglasID}", func(w http.ResponseWriter, r *http.Request) {
		handleDodajKomentar(mongoClient, w, r)
	})

	mux.HandleFunc("/api/obrisiKomentar/{komentarUUID}", func(w http.ResponseWriter, r *http.Request) {
		handleObrisiKomentar(mongoClient, w, r)
	})

	mux.HandleFunc("/noviOglas/{sport}", func(w http.ResponseWriter, r *http.Request) {
		handleNoviOglas(mongoClient, w, r)
	})

	mux.HandleFunc("/api/dodajOglas", func(w http.ResponseWriter, r *http.Request) {
		handleDodajOglas(mongoClient, w, r)
	})

	mux.HandleFunc("/profil/{username}", func(w http.ResponseWriter, r *http.Request) {
		handleProfil(mongoClient, w, r)
	})

	mux.HandleFunc("/api/oceniKorisnika/{korisnikUsername}", func(w http.ResponseWriter, r *http.Request) {
		handleOceniKorisnika(mongoClient, w, r)
	})

	mux.HandleFunc("/promeniSlikuProfila", func(w http.ResponseWriter, r *http.Request) {
		handlePromeniSlikuProfilaDialog(mongoClient, w, r)
	})

	mux.HandleFunc("/api/promeniSlikuProfila", func(w http.ResponseWriter, r *http.Request) {
		handlePromeniSlikuProfila(mongoClient, w, r)
	})

	mux.HandleFunc("/api/obrisiNalog/{username}", func(w http.ResponseWriter, r *http.Request) {
		handleObrisiNalog(mongoClient, w, r)
	})

	mux.HandleFunc("/api/izmeniOglas/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleIzmeniOglas(mongoClient, w, r)
	})

	mux.HandleFunc("/api/obrisiOglas/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleObrisiOglas(mongoClient, w, r)
	})

	mux.HandleFunc("/NotFound", func(w http.ResponseWriter, r *http.Request) {
		handleNotFound(mongoClient, w, r)
	})

	return mux
}

func handleHome(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.Header().Set("HX-Redirect", "/NotFound")
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		views.PocetnaStranica(nil).Render(r.Context(), w)
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		views.PocetnaStranica(nil).Render(r.Context(), w)
		return
	}

	korisnik, err := db.VratiKorisnika(mongoClient, claims.Username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
	}

	views.PocetnaStranica(korisnik).Render(r.Context(), w)
}

func handleRegistrujKorisnika(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ime := r.FormValue("signup-ime")
	prezime := r.FormValue("signup-prezime")
	username := r.FormValue("signup-username")
	lozinka := r.FormValue("signup-lozinka")
	ponoviLozinku := r.FormValue("signup-ponovi-lozinku")

	if !validanUsername(username) {
		log.Printf("Greška prilikom registrovanja korisnika %v: nevalidan username\n", username)
		views.SignupGreska("Korisničko ime mora da ima najmanje 5 karaktera i može da sadrži slova, brojeve i znak '_'!").Render(r.Context(), w)
		return
	}

	vracenKorisnik, err := db.VratiKorisnika(mongoClient, username)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("Greška prilikom registrovanja korisnika %v: %v\n", username, err)
			views.SignupGreska("Interna serverska greška!").Render(r.Context(), w)
			return
		}
	}

	if vracenKorisnik != nil {
		log.Printf("Greška prilikom registrovanja korisnika: korisnik %v već postoji!\n", username)
		views.SignupGreska(fmt.Sprintf("Korisnik %v već postoji!", username)).Render(r.Context(), w)
		return
	}

	if !validnaLozinka(lozinka) {
		log.Printf("Greška prilikom registrovanja korisnika %v: nevalidna lozinka\n", username)
		poruka := "Lozinka mora da ima najmanje 8 karaktera, 1 malo slovo, 1 veliko slovo i 1 specijalni karakter!"
		views.SignupGreska(poruka).Render(r.Context(), w)
		return
	}

	if lozinka == "" || lozinka != ponoviLozinku {
		views.SignupGreska("Lozinke se ne poklapaju!").Render(r.Context(), w)
		log.Printf("Greška prilikom registrovanja korisnika %v: lozinke se ne poklapaju\n", username)
		return
	}

	hes, err := bcrypt.GenerateFromPassword([]byte(lozinka), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Greška prilikom registrovanja korisnika %v: %v\n", username, err)
		views.SignupGreska("Interna serverska greška!").Render(r.Context(), w)
		return
	}

	lozinkaHes := string(hes)

	korisnik := db.Korisnik {
		Ime: ime,
		Prezime: prezime,
		Username: username,
		LozinkaHes: lozinkaHes,
		SlikaBase64: "data:image/jpg;base64, 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",
	}

	token, err := generisiJWT(username)
	if err != nil {
		log.Printf("Greška prilikom registrovanja korisnika %v: %v\n", username, err)
		views.SignupGreska("Interna serverska greška!").Render(r.Context(), w)
		return
	}

	_, err = db.DodajKorisnika(mongoClient, korisnik)
	if err != nil {
		views.SignupGreska(fmt.Sprintf("Korisnik %v već postoji!", username)).Render(r.Context(), w)
		log.Printf("Greška prilikom registrovanja korisnika %v: %v\n", username, err)
		return
	}

	http.SetCookie(w, &http.Cookie {
		Name: "sportify_session_token",
		Value: token,
		HttpOnly: true,
		Path: "/",
	})

	w.Header().Set("HX-Redirect", "/")

    log.Printf("Korisnik %v uspešno registrovan\n", username)
}

func handlePrijaviKorisnika(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("login-username")
	lozinka := r.FormValue("login-lozinka")

	vracenKorisnik, err := db.VratiKorisnika(mongoClient, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("Greška pirlikom prijavljivanja korisnika %v: Ne postoji dati korisnik")
			views.LoginGreska("Korisničko ime ili lozinka nisu tačni!").Render(r.Context(), w)
			return
		} else {
			log.Printf("Greška prilikom prijavljivanja korisnika %v: %v\n", username, err)
			views.LoginGreska("Interna serverska greška!").Render(r.Context(), w)
			return
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(vracenKorisnik.LozinkaHes), []byte(lozinka)); err != nil {
		views.LoginGreska("Korisničko ime ili lozinka nisu tačni!").Render(r.Context(), w)
		return
	}

	token, err := generisiJWT(username)
	if err != nil {
		log.Printf("Greška prilikom prijavljivanja korisnika %v: %v\n", username, err)
		views.LoginGreska("Interna serverska greška!").Render(r.Context(), w)
		return
	}

	http.SetCookie(w, &http.Cookie {
		Name: "sportify_session_token",
		Value: token,
		HttpOnly: true,
		Path: "/",
	})

	w.Header().Set("HX-Redirect", "/")

    log.Printf("Korisnik %v uspešno prijavljen\n", username)
}

func handleOdjaviKorisnika(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie {
		Name: "sportify_session_token",
		Value: "",
		HttpOnly: true,
		Path: "/",
		MaxAge: -1,
	})

	w.Header().Set("HX-Redirect", "/")
}

func handleSviOglasi(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stranaStr := r.URL.Query().Get("strana")
	oglasaPoStraniStr := r.URL.Query().Get("brojOglasa")

	strana, err := strconv.Atoi(stranaStr)
	if err != nil || strana < 1 {
		strana = 1
	}

	oglasaPoStrani, err := strconv.Atoi(oglasaPoStraniStr)
	if err != nil || oglasaPoStrani < 5 || oglasaPoStrani > 20 {
		oglasaPoStrani = 10
	}

	oglasi, err := db.SviOglasi(mongoClient)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja svih oglasa: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	sportovi, err := db.SviSportovi(mongoClient)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja svih sportova: %v\n", err)
		w.Header().Set("HX-Redirect", "/")

		return
	}

	lokacije, err := db.SveLokacije(mongoClient)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja svih lokacija: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	ukupnoStrana := int(math.Ceil(float64(len(oglasi)) / float64(oglasaPoStrani)))
	idxPrvogOglasa := (strana - 1) * oglasaPoStrani
	idxPoslednjegOglasa := idxPrvogOglasa + oglasaPoStrani

	if idxPrvogOglasa >= len(oglasi) {
		oglasi = make([]db.Oglas, 0)
		idxPrvogOglasa = 0
	}

	if idxPoslednjegOglasa > len(oglasi) {
		idxPoslednjegOglasa = len(oglasi)
	}


	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		views.SviOglasi(nil, oglasi[idxPrvogOglasa:idxPoslednjegOglasa], sportovi, lokacije, strana, ukupnoStrana, oglasaPoStrani).Render(r.Context(), w)
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		views.SviOglasi(nil, oglasi[idxPrvogOglasa:idxPoslednjegOglasa], sportovi, lokacije, strana, ukupnoStrana, oglasaPoStrani).Render(r.Context(), w)
		return
	}

	korisnik, err := db.VratiKorisnika(mongoClient, claims.Username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
	}

	views.SviOglasi(korisnik, oglasi[idxPrvogOglasa:idxPoslednjegOglasa], sportovi, lokacije, strana, ukupnoStrana, oglasaPoStrani).Render(r.Context(), w)
}

func handleOglasiFilter(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sviOglasi, err := db.SviOglasi(mongoClient)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja svih oglasa: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Printf("Greška prilikom parsovanja forme: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	sportovi := make([]string, 0)
	lokacije := make([]string, 0)
	var (
		datumOd time.Time
		datumDo time.Time
		nilDatum time.Time
	)

	for k, v := range r.Form {
		switch k[13:18] {
			case "sport":
				sportovi = append(sportovi, k[19:])

			case "datum":
				if k[19:21] == "od" {
					datumOd, _ = time.Parse(time.DateOnly, v[0])
				} else {
					datumDo, _ = time.Parse(time.DateOnly, v[0])
				}

			default:
				lokacije = append(lokacije, k[22:])
		}
	}

	if len(sportovi) == 0 && len(lokacije) == 0 && datumOd.Equal(nilDatum) && datumDo.Equal(nilDatum) {
		views.FiltriraniOglasi(sviOglasi).Render(r.Context(), w)
		return
	}

	oglasi := make([]db.Oglas, 0)
	if !datumOd.Equal(datumDo) {
		if datumDo.Equal(nilDatum) {
			datumDo = time.Now().AddDate(100, 0, 0)
		}

		for _, oglas := range sviOglasi {
			if oglas.Vreme.Before(datumDo) && oglas.Vreme.After(datumOd) {
				oglasi = append(oglasi, oglas)
			}
		}
	} else {
		if datumOd.Equal(nilDatum) {
			oglasi = append(oglasi, sviOglasi...)
		} else {
			for _, oglas := range sviOglasi {
				oglasG, oglasM, oglasD := oglas.Vreme.Date()
				datumOdG, datumOdM, datumOdD := datumOd.Date()

				if oglasG == datumOdG && oglasM == datumOdM && oglasD == datumOdD {
					oglasi = append(oglasi, oglas)
				}
			}
		}
	}

	filtriraniOglasi := make([]db.Oglas, 0)
	for _, oglas := range oglasi {
		if slices.Contains(sportovi, oglas.Sport) {
			filtriraniOglasi = append(filtriraniOglasi, oglas)
		}

		if slices.Contains(lokacije, oglas.Mesto) {
			filtriraniOglasi = append(filtriraniOglasi, oglas)
		}
	}

	views.FiltriraniOglasi(filtriraniOglasi).Render(r.Context(), w)
}

func handlePrikazOglasa(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hexID := r.PathValue("id")
	if hexID == "" {
		log.Printf("Greška prilikom pribavljanja id-a oglasa preko putanje.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	objectID, err := bson.ObjectIDFromHex(hexID)
	if err != nil {
		log.Printf("Greška prilikom kreiranja ObjectID-a: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	oglas, err := db.VratiOglas(mongoClient, objectID)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja oglasa: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	var korisnik *db.Korisnik
	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
			w.Header().Set("HX-Redirect", "/oglasi")
			return
		}
	} else {

		claims, err := parseJWT(cookie.Value)
		if err != nil {
			log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
			w.Header().Set("HX-Redirect", "/oglasi")
			return
		}

		korisnik, err = db.VratiKorisnika(mongoClient, claims.Username)
		if err != nil {
			log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
		}
	}

	views.PrikazOglasa(korisnik, oglas).Render(r.Context(), w)
}

func handleDodajKomentar(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vlasnik := r.FormValue("vlasnik")
	if vlasnik == "" {
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	sadrzaj := r.FormValue("komentar-sadrzaj")
	if sadrzaj == "" {
		w.Header().Set("HX-Reswap", "none")
		return
	}

	hexID := r.PathValue("oglasID")
	if hexID == "" {
		log.Printf("Greška prilikom pribavljanja id-a objekta preko putanje.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	objectID, err := bson.ObjectIDFromHex(hexID)
	if err != nil {
		log.Printf("Greška prilikom kreiranja ObjectID-a: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	oglas, err := db.VratiOglas(mongoClient, objectID)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja oglasa: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	komentar := db.Komentar {
		UUID: uuid.NewString(),
		Vlasnik: vlasnik,
		Sadrzaj: sadrzaj,
		Vreme: time.Now(),
	}

	if err := db.DodajKomentar(mongoClient, objectID, komentar); err != nil {
		log.Printf("Greška prilikom dodavanja komentara: %v\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	if len(oglas.Komentari) == 0 {
		w.Header().Set("HX-Reswap", "innerHTML")
	} else {
		w.Header().Set("HX-Reswap", "beforeend")
	}

	views.CrtajKomentar(komentar, vlasnik, oglas.ID.Hex()).Render(r.Context(), w)
}

func handleObrisiKomentar(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	komentarUUID := r.PathValue("komentarUUID")
	if komentarUUID == "" {
		log.Printf("Greška prilikom pribavljanja id-a komentara.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	oglasID := r.FormValue("oglasID")
	if oglasID == "" {
		log.Printf("Greška prilikom pribavljanja id-a oglasa.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	objectID, err := bson.ObjectIDFromHex(oglasID)
	if err != nil {
		log.Printf("Greška prilikom kreiranja ObjectID-a: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	oglas, err := db.VratiOglas(mongoClient, objectID)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja oglasa: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	if err := db.ObrisiKomentar(mongoClient, objectID, komentarUUID); err != nil {
		log.Printf("Greška prilikom dodavanja komentara: %v\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	if len(oglas.Komentari) == 1 {
		views.NemaKomentara().Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Reswap", "delete")
}

func handleNoviOglas(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nazivSporta := r.PathValue("sport")

	var (
		sport *db.Sport
		err error
	)

	if nazivSporta != "*" {
		sport, err = db.VratiSport(mongoClient, nazivSporta)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments){
			log.Printf("Greška prilikom pribavljanja sporta: %v\n", err)
			w.Header().Set("HX-Redirect", "/")
			return
		}
	}

	sportovi, err := db.SviSportovi(mongoClient)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja svih sportova: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	lokacije, err := db.SveLokacije(mongoClient)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja svih lokacija: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	korisnik, err := db.VratiKorisnika(mongoClient, claims.Username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	views.NoviOglas(korisnik, sportovi, lokacije, sport).Render(r.Context(), w)
}

func handleDodajOglas(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sport := r.FormValue("novi-oglas-sport")
	mesto := r.FormValue("novi-oglas-mesto")
	datumStr := r.FormValue("novi-oglas-datum")
	vremeStr := r.FormValue("novi-oglas-vreme")
	potrebnoIgracaStr := r.FormValue("novi-oglas-potrebno-igraca")
	opis := r.FormValue("novi-oglas-opis")
	koordinateStr := r.FormValue("novi-oglas-koordinate")

	datum, err := time.Parse(time.DateOnly, datumStr)
	if err != nil {
		log.Printf("Greška prilikom parsiranja datuma: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	vreme, err := time.Parse(time.TimeOnly, vremeStr + ":00")
	if err != nil {
		log.Printf("Greška prilikom parsiranja vremena: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	godina, mesec, dan := datum.Date()
	vreme = vreme.AddDate(godina, int(mesec) - 1, dan - 1)

	potrebnoIgraca, err := strconv.Atoi(potrebnoIgracaStr)
	if err != nil {
		log.Printf("Greška prilikom parsiranja potrebnog broja igrača: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	koordinate := [2]float64 {0, 0}
	if koordinateStr != "" {
		coords := strings.Split(koordinateStr, ",")
		for i, koordinata := range coords {
			broj, err := strconv.ParseFloat(koordinata, 64)
			if err != nil {
				log.Printf("Greška prilikom parsiranja koordinate %v: %v\n", koordinata, err)
				w.Header().Set("HX-Redirect", "/")
				return
			}

			koordinate[i] = broj
		}
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	oglas := db.Oglas {
		Vlasnik: claims.Username,
		Sport: sport,
		Opis: opis,
		PotrebnoIgraca: potrebnoIgraca,
		Mesto: mesto,
		Koordinate: koordinate,
		Vreme: vreme,
	}

	oglasID, err := db.DodajOglas(mongoClient, oglas)
	if err != nil {
		log.Printf("Greška prilikom dodavanja novog oglasa: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	log.Printf("Uspešno dodat oglas: %v\n", oglasID.Hex())
	w.Header().Set("HX-Redirect", "/oglasi")
}

func handleProfil(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.PathValue("username")
	if username == "" {
		log.Printf("Greška prilikom pribavljanja username-a preko putanje.\n")
		w.Header().Set("HX-Redirect", "/")
		return
	}

	var korisnik *db.Korisnik
	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
			w.Header().Set("HX-Redirect", "/")
			return
		}
	} else {

		claims, err := parseJWT(cookie.Value)
		if err != nil {
			log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
			w.Header().Set("HX-Redirect", "/")
			return
		}

		korisnik, err = db.VratiKorisnika(mongoClient, claims.Username)
		if err != nil {
			log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
		}
	}

	korisnikPrikaz, err := db.VratiKorisnika(mongoClient, username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", username, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	oglasi, err := db.SviOglasiKorisnika(mongoClient, username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja svih oglasa korisnika %v: %v\n", username, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	var ocena *db.Ocena
	if korisnik != nil {
		ocena, err = db.VratiOcenuKorisnika(mongoClient, username, korisnik.Username)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("Greška prilikom pribavljanja ocene korisnika %v: %v\n", username, err)
			w.Header().Set("HX-Redirect", "/")
			return
		}
	}

	views.Profil(korisnik, korisnikPrikaz, oglasi, ocena).Render(r.Context(), w)
}

func handleOceniKorisnika(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	korisnikUsername := r.PathValue("korisnikUsername")
	ocenaStr := r.FormValue("ocena")

	ocenaVrednost, err := strconv.Atoi(ocenaStr)
	if err != nil {
		log.Printf("Greška prilikom parsiranja ocene: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	ucitanaOcena, err := db.VratiOcenuKorisnika(mongoClient, korisnikUsername, claims.Username)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("Greška prilikom pribavljanja ocene korisnika %v: %v\n", korisnikUsername, err)
			w.Header().Set("HX-Redirect", "/")
			return
		}

		ocena := db.Ocena {
			UUID: uuid.NewString(),
			Vlasnik: claims.Username,
			Vrednost: uint(ocenaVrednost),
		}

		if err = db.DodajOcenuKorisnika(mongoClient, korisnikUsername, ocena); err != nil {
			log.Printf("Greška prilikom dodavanja ocene korisnika %v: %v\n", korisnikUsername, err)
			w.Header().Set("HX-Redirect", "/")
			return
		}

		w.Header().Set("HX-Refresh", "true")
		return
	}

	ucitanaOcena.Vrednost = uint(ocenaVrednost)
	if err = db.IzmeniOcenuKorisnika(mongoClient, korisnikUsername, *ucitanaOcena); err != nil {
		log.Printf("Greška prilikom izmene ocene korisnika %v: %v\n", korisnikUsername, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	w.Header().Set("HX-Refresh", "true")
}

func handlePromeniSlikuProfilaDialog(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	korisnik, err := db.VratiKorisnika(mongoClient, claims.Username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	views.PromeniSlikuProfila(korisnik).Render(r.Context(), w)
}

func handlePromeniSlikuProfila(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slikaBase64 := r.FormValue("nova-slika-base64")
	if slikaBase64 == "" {
		log.Printf("Greška prilikom pribavljanja base64 stringa slike\n")
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	korisnik, err := db.VratiKorisnika(mongoClient, claims.Username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	korisnik.SlikaBase64 = slikaBase64

	if err := db.IzmeniKorisnika(mongoClient, *korisnik); err != nil {
		log.Printf("Greška prilikom izmene korisnika %v: %v\n", claims.Username, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	w.Header().Set("HX-Refresh", "true")
}

func handleObrisiNalog(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.PathValue("username")

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	if username != claims.Username {
		log.Printf("Greška prilikom brisanja korisnika %v: nalozi %v i %v nisu isti!\n", username, username, claims.Username)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	brojObrisanih, err := db.ObrisiKorisnika(mongoClient, username)
	if err != nil {
		log.Printf("Greška prilikom brisanja korisnika %v: %v\n", username, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	if brojObrisanih != 0 {
		http.SetCookie(w, &http.Cookie {
			Name: "sportify_session_token",
			Value: "",
			HttpOnly: true,
			Path: "/",
			MaxAge: -1,
		})

		w.Header().Set("HX-Redirect", "/")
	}

	log.Printf("Korisnik %v uspešno obrisan.\n", username)
}


func handleIzmeniOglas(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hexID := r.PathValue("id")
	if hexID == "" {
		log.Printf("Greška prilikom pribavljanja id-a oglasa preko putanje.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	objectID, err := bson.ObjectIDFromHex(hexID)
	if err != nil {
		log.Printf("Greška prilikom kreiranja ObjectID-a: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	oglas, err := db.VratiOglas(mongoClient, objectID)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja oglasa: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	if oglas.Vlasnik != claims.Username {
		log.Printf("Greška prilikom izmene oglasa sa ID %v: korisnik %v nije vlasnik oglasa!\n", hexID, claims.Username)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	koordinateStr := r.FormValue("izmeni-oglas-koordinate")
	potrebnoIgracaStr := r.FormValue("izmeni-potrebno-igraca")

	if koordinateStr == "" {
		log.Printf("Greška prilikom pribavljanja koordinata.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	if potrebnoIgracaStr == "" {
		log.Printf("Greška prilikom pribavljanja broja potrebnih igraca.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	koordinate := [2]float64 {0, 0}
	if koordinateStr != "" {
		coords := strings.Split(koordinateStr, ",")
		for i, koordinata := range coords {
			broj, err := strconv.ParseFloat(koordinata, 64)
			if err != nil {
				log.Printf("Greška prilikom parsiranja koordinate %v: %v\n", koordinata, err)
				w.Header().Set("HX-Redirect", "/")
				return
			}

			koordinate[i] = broj
		}
	}

	potrebnoIgraca, err := strconv.Atoi(potrebnoIgracaStr)
	if err != nil {
		log.Printf("Greška prilikom parsiranja potrebnog broja igrača: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	oglas.Koordinate = koordinate
	oglas.PotrebnoIgraca = potrebnoIgraca

	if err := db.IzmeniOglas(mongoClient, *oglas); err != nil {
		log.Printf("Greška prilikom izmene oglasa: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	w.Header().Set("HX-Refresh", "true")
}

func handleObrisiOglas(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hexID := r.PathValue("id")
	if hexID == "" {
		log.Printf("Greška prilikom pribavljanja id-a oglasa preko putanje.\n")
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	objectID, err := bson.ObjectIDFromHex(hexID)
	if err != nil {
		log.Printf("Greška prilikom kreiranja ObjectID-a: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	oglas, err := db.VratiOglas(mongoClient, objectID)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja oglasa: %v.\n", err)
		w.Header().Set("HX-Redirect", "/oglasi")
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	if oglas.Vlasnik != claims.Username {
		log.Printf("Greška prilikom brisanja oglasa sa ID %v: korisnik %v nije vlasnik oglasa!\n", hexID, claims.Username)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	brojObrisanih, err := db.ObrisiOglas(mongoClient, objectID)
	if err != nil {
		log.Printf("Greška prilikom brisanja oglasa sa ID %v: %v\n", hexID, err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	if brojObrisanih != 0 {
		log.Printf("Uspešno obrisan oglas sa ID: %v\n", hexID)
	}

	w.Header().Set("HX-Redirect", "/oglasi")
}

func handleNotFound(mongoClient *mongo.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("sportify_session_token")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Printf("Greška prilikom učitavanja session tokena: %v\n", err)
		}
		w.Header().Set("HX-Redirect", "/")
		return
	}

	claims, err := parseJWT(cookie.Value)
	if err != nil {
		log.Printf("Greška prilikom parsiranja session tokena: %v\n", err)
		w.Header().Set("HX-Redirect", "/")
		return
	}

	korisnik, err := db.VratiKorisnika(mongoClient, claims.Username)
	if err != nil {
		log.Printf("Greška prilikom pribavljanja korisnika %v: %v\n", claims.Username, err)
	}

	views.NotFound(korisnik).Render(r.Context(), w)
}

func NoviServer(mongoClient *mongo.Client) (*http.Server, int) {
	noviServer := Server {
		port: 8080,
	}

	return &http.Server {
		Addr: fmt.Sprintf(":%d", noviServer.port),
		Handler: noviServer.HandlerInit(mongoClient),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}, noviServer.port
}

func validanUsername(username string) bool {
	if len(username) < 5 {
		return false
	}

	for _, char := range username {
		if char >= 'a' && char <= 'z' {
			continue
		}

		if char >= '0' && char <= '9' {
			continue
		}

		if char == '_' {
			continue
		}

		return false
	}

	return true
}

func validnaLozinka(lozinka string) bool {
	if len(lozinka) < 8 {
		return false
	}

	specKarakter, maloSlovo, velikoSlovo, broj := false, false, false, false

	for _, char := range lozinka {
		if !specKarakter && slices.Contains(lozinkaSpecKarakteri, char) {
			specKarakter = true
		}

		if char >= 'a' && char <= 'z' {
			maloSlovo = true
		}

		if char >= 'A' && char <= 'Z' {
			velikoSlovo = true
		}

		if char >= '0' && char <= '9' {
			broj = true
		}

		if specKarakter && maloSlovo && velikoSlovo && broj {
			return true
		}
	}

	return false
}

func generisiJWT(username string) (string, error) {
	claims := KorisnikClaims {
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtKljuc))
}

func parseJWT(tokenString string) (*KorisnikClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &KorisnikClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Neočekivani algoritam: %v", token.Header["alg"])
		}

		return []byte(jwtKljuc), nil
	})

	if err != nil {
		return nil, fmt.Errorf("Nevažeći token: %v", err)
	}

	claims, ok := token.Claims.(*KorisnikClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Nije moguće parsirati claims")
	}

	return claims, nil
}
