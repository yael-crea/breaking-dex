# Breaking Dex — DataDex : fetch & show

Site web en Go qui exploite l'API REST gratuite **TVmaze** pour présenter l'univers de la série **Breaking Bad** : ses saisons, ses 62 épisodes et ses personnages.

## Présentation

- **Contexte** : projet individuel *DataDex : fetch & show*. L'objectif est de récupérer des données JSON depuis une API REST et de les manipuler à la main en Go.
- **Thème** : la série Breaking Bad (AMC, 2008-2013).
- **Technologies** : Go (bibliothèque standard uniquement), HTML (`html/template`) et CSS. Le site n'utilise aucun JavaScript.

### Fonctionnalités

| | Fonctionnalité | Détail |
|---|---|---|
| FT1 | Recherche | Sur les épisodes (titre **et** résumé) et sur les personnages (nom du personnage **et** nom de l'acteur). Insensible à la casse et aux accents. |
| FT2 | Filtres cumulables | Sur les épisodes : **saison**, **année de diffusion**, **note minimale** et **durée maximale**. Tous les filtres se combinent. |
| FT3 | Pagination | 10, 20 ou 30 éléments par page, au choix de l'utilisateur. Elle est disponible sur la collection et sur la recherche, et les filtres sont conservés d'une page à l'autre. |
| FT4 | Favoris | Ajout, suppression et vidage de la liste. Elle est enregistrée dans `favoris.json` côté serveur et commune à tous les visiteurs. |
| Bonus | | Tri (diffusion, note, titre), épisode précédent / suivant, acteurs invités, cache mémoire des données de l'API, design adapté au mobile. |

La recherche, les filtres, le tri et la pagination sont codés à la main dans `collection.go` : le programme parcourt lui-même la collection récupérée et n'utilise aucun paramètre de recherche, de filtre ou de pagination de l'API.

### Gestion des erreurs

- Si l'API est hors ligne ou ne répond pas dans les 10 secondes, le site affiche une page **503** et ne plante pas. Si des données ont déjà été chargées, elles sont réutilisées.
- Une ressource inexistante (épisode, acteur, route) affiche une page **404**.
- Une valeur invalide dans l'URL (`?saison=abc`, `?par_page=15`…) affiche une page **400** avec un message clair.
- Un numéro de page trop grand est ramené à la dernière page.
- Une erreur de lecture ou d'écriture du fichier des favoris affiche une page **500**.

## Installation et lancement

### Prérequis

- [Go](https://go.dev/dl/) **1.22 ou plus récent**, nécessaire pour les routes du type `GET /episodes/{id}`. Le projet a été développé avec Go 1.27.
- Une connexion Internet, pour joindre l'API TVmaze.
- Aucune clé d'API n'est nécessaire.

### Lancer le projet

```bash
git clone <url-du-depot>
cd breaking-bad
go run .
```

Ouvrir ensuite <http://localhost:8080>.

Pour changer de port :

```bash
PORT=3000 go run .            # Linux / macOS
$env:PORT="3000"; go run .    # Windows PowerShell
```

Le fichier `favoris.json` est créé automatiquement au premier ajout d'un favori.

## Routes

| Méthode | Route | Description |
|---|---|---|
| GET | `/` | Accueil : présentation de la série, saisons, épisodes les mieux notés, personnages principaux |
| GET | `/episodes` | Collection des épisodes avec filtres, tri et pagination. Paramètres : `saison`, `annee`, `note`, `duree`, `tri`, `page`, `par_page` |
| GET | `/episodes/{id}` | Détail d'un épisode : image, fiche, résumé, acteurs invités, épisodes précédent et suivant |
| GET | `/personnages` | Liste des personnages principaux et de leurs acteurs |
| GET | `/personnages/{id}` | Fiche d'un acteur (`id` = identifiant TVmaze de l'acteur) |
| GET | `/recherche` | Résultats de recherche. Paramètres : `q`, `page`, `par_page` |
| GET | `/favoris` | Liste des épisodes favoris |
| POST | `/favoris/ajouter` | Ajoute un épisode aux favoris (champs `id`, `retour`) |
| POST | `/favoris/supprimer` | Retire un épisode des favoris (champs `id`, `retour`) |
| POST | `/favoris/vider` | Vide la liste des favoris |
| GET | `/a-propos` | Présentation du projet et retour d'expérience (FAQ) |
| GET | `/static/...` | Fichiers statiques (CSS) |
| * | toute autre route | Page d'erreur 404 |

## API utilisée : TVmaze

- Documentation : <https://www.tvmaze.com/api>
- URL de base : `https://api.tvmaze.com`
- API REST, réponses JSON, gratuite, sans clé ni authentification.
- Limite d'environ 20 requêtes toutes les 10 secondes par adresse IP. Le cache mémoire de 10 minutes du site permet de rester bien en dessous.
- L'identifiant de Breaking Bad sur TVmaze est `169`.
- L'API ne renvoie ses textes qu'en anglais. Le site affiche donc des résumés traduits en français, stockés dans `donnees/traductions_fr.json` (clé = identifiant de l'épisode). Les genres et les pays sont traduits par de petits dictionnaires dans `traduction.go`. Si une traduction manque, le texte anglais de l'API est affiché. Les titres des épisodes restent les titres originaux.

### Endpoints exploités

| Endpoint | Utilisation dans le site |
|---|---|
| `GET /shows/169` | Informations générales de la série : nom, genres, chaîne, note, affiche, résumé. Utilisé sur la page d'accueil. |
| `GET /shows/169/seasons` | Les 5 saisons : nombre d'épisodes, dates, affiches. Utilisé sur l'accueil et dans le filtre par saison. |
| `GET /shows/169/episodes` | La collection des 62 épisodes. Base de la collection, de la recherche, des filtres, de la pagination et des favoris. |
| `GET /shows/169/cast` | Le casting principal (personnages et acteurs). Utilisé sur la page personnages et dans la recherche. |
| `GET /episodes/{id}?embed=guestcast` | Le détail d'un épisode avec ses acteurs invités. Utilisé sur la page ressource. |
| `GET /people/{id}` | La fiche d'un acteur : naissance, décès, genre, pays, photo. Utilisé sur la page de détail d'un personnage. |

## Organisation du code

```
breaking-bad/
├── main.go          # démarrage du serveur, chargement des templates, routes
├── api.go           # structures JSON, appels HTTP à TVmaze, cache mémoire
├── collection.go    # recherche, filtres, tri et pagination (faits à la main)
├── favoris.go       # lecture / écriture de favoris.json
├── commun.go        # données communes, affichage des templates, pages d'erreur
├── page_*.go        # un fichier par page (accueil, episodes, recherche, personnages, favoris, apropos)
├── traduction.go    # chargement des traductions françaises, genres et pays
├── donnees/traductions_fr.json  # résumés de la série et des 62 épisodes en français
├── templates/       # base.html (header, footer, composants) + une page par fichier
└── static/style.css
```
