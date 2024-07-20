port module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes as Attr
import Html.Events exposing (..)
import Html.Parser
import Html.Parser.Util
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import List
import Markdown.Parser as Markdown
import Markdown.Renderer
import String
import Svg
import Svg.Attributes
import Time
import Url


port turnstileToken : (String -> msg) -> Sub msg


urlNewComment : String
urlNewComment =
    "https://api.carlo-hamalainen.net/v1/comments/new"


urlComments : String
urlComments =
    "https://api.carlo-hamalainen.net/v1/comments"


theSiteId : String
theSiteId =
    "carlo-hamalainen.net"

spinner : Html msg
spinner =
    Svg.svg
        [ Svg.Attributes.width "50"
        , Svg.Attributes.height "50"
        , Svg.Attributes.viewBox "0 0 50 50"
        ]
        [ Svg.circle
            [ Svg.Attributes.cx "25"
            , Svg.Attributes.cy "25"
            , Svg.Attributes.r "20"
            , Svg.Attributes.fill "none"
            , Svg.Attributes.stroke "#007bff"
            , Svg.Attributes.strokeWidth "5"
            , Svg.Attributes.strokeLinecap "round"
            , Svg.Attributes.strokeDasharray "80"
            , Svg.Attributes.strokeDashoffset "60"
            ]
            [ Svg.animateTransform
                [ Svg.Attributes.attributeName "transform"
                , Svg.Attributes.type_ "rotate"
                , Svg.Attributes.from "0 25 25"
                , Svg.Attributes.to "360 25 25"
                , Svg.Attributes.dur "1s"
                , Svg.Attributes.repeatCount "indefinite"
                ]
                []
            ]
        ]


type alias Comment t =
    { commentID : String
    , siteID : String
    , postID : String
    , timestamp : t
    , author : String
    , authorEmail : String
    , commentBody : String
    , isActive : Bool
    }


type alias LoadedComments =
    { allComments : List (Comment Time.Posix)
    , unmoderated : List (Comment Time.Posix)
    }


type State a
    = Failure String
    | Idle
    | Loading
    | Success a


type alias Model =
    { lc : State LoadedComments
    , currentUrl : Url.Url
    , author : String
    , authorEmail : String
    , commentBody : String
    , submitStatus : SubmitStatus
    , turnstileToken : Maybe String
    }


type SubmitStatus
    = NotSubmitted
    | Submitting
    | SubmitSuccess
    | SubmitFailure String


type alias Flags =
    { url : String }


main : Program Flags Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }


subscriptions : Model -> Sub Msg
subscriptions _ =
    turnstileToken TurnstileToken


defaultUrl : Url.Url
defaultUrl =
    { protocol = Url.Http
    , host = "example.com"
    , port_ = Nothing
    , path = "/"
    , query = Nothing
    , fragment = Nothing
    }


init : Flags -> ( Model, Cmd Msg )
init flags =
    let
        url =
            Maybe.withDefault defaultUrl (Url.fromString flags.url)
    in
    ( { author = ""
      , authorEmail = ""
      , commentBody = ""
      , submitStatus = NotSubmitted
      , lc = Loading
      , currentUrl = url
      , turnstileToken = Nothing
      }
    , getComments url
    )


type Msg
    = GotComments (Result Http.Error (List (Comment Int)))
    | UpdateAuthor String
    | UpdateAuthorEmail String
    | UpdateCommentBody String
    | SubmitForm
    | GotSubmitResponse (Result Http.Error ())
    | TurnstileToken String


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UpdateAuthor newAuthor ->
            ( { model | author = newAuthor }, Cmd.none )

        UpdateAuthorEmail newEmail ->
            ( { model | authorEmail = newEmail }, Cmd.none )

        UpdateCommentBody newBody ->
            ( { model | commentBody = newBody }, Cmd.none )

        GotComments result ->
            case result of
                Ok data ->
                    let
                        parsedComments =
                            List.map
                                (\comment ->
                                    { commentID = comment.commentID
                                    , siteID = comment.siteID
                                    , postID = comment.postID
                                    , author = comment.author
                                    , authorEmail = comment.authorEmail
                                    , timestamp = Time.millisToPosix comment.timestamp
                                    , commentBody = comment.commentBody
                                    , isActive = comment.isActive
                                    }
                                )
                                data
                    in
                    ( { model | lc = Success { allComments = parsedComments, unmoderated = [] } }, Cmd.none )

                Err e ->
                    ( { model | lc = Failure (errorToString e) }, Cmd.none )

        SubmitForm ->
            ( { model | submitStatus = Submitting }
            , submitComment model
            )

        GotSubmitResponse result ->
            case result of
                Ok _ ->
                    ( { model | submitStatus = SubmitSuccess }, Cmd.none )

                Err error ->
                    ( { model | submitStatus = SubmitFailure (errorToString error) }, Cmd.none )

        TurnstileToken token ->
            ( { model | turnstileToken = Just token }, Cmd.none )


updateGotComments : Model -> List (Comment Int) -> ( Model, Cmd Msg )
updateGotComments model data =
    let
        parsedComments =
            List.map
                (\comment ->
                    { commentID = comment.commentID
                    , siteID = comment.siteID
                    , postID = comment.postID
                    , author = comment.author
                    , authorEmail = comment.authorEmail
                    , timestamp = Time.millisToPosix comment.timestamp
                    , commentBody = comment.commentBody
                    , isActive = comment.isActive
                    }
                )
                data
    in
    ( { model | lc = Success { allComments = parsedComments, unmoderated = [] } }, Cmd.none )


submitComment : Model -> Cmd Msg
submitComment model =
    Http.post
        { url = urlNewComment
        , body = Http.jsonBody (encodeComment model)
        , expect = Http.expectWhatever GotSubmitResponse
        }


encodeComment : Model -> Encode.Value
encodeComment model =
    Encode.object
        [ ( "siteID", Encode.string theSiteId )
        , ( "postID", Encode.string (trimTrailingSlashes model.currentUrl.path) )
        , ( "author", Encode.string model.author )
        , ( "authorEmail", Encode.string model.authorEmail )
        , ( "commentBody", Encode.string model.commentBody )
        , ( "turnstileToken", Encode.string (Maybe.withDefault "" model.turnstileToken) )
        ]


errorToString : Http.Error -> String
errorToString error =
    case error of
        Http.BadUrl url ->
            "Bad URL: " ++ url

        Http.Timeout ->
            "Request timed out"

        Http.NetworkError ->
            "Network error"

        Http.BadStatus statusCode ->
            "Bad status: " ++ String.fromInt statusCode

        Http.BadBody message ->
            "Bad body: " ++ message


httpErrorToString : Http.Error -> String
httpErrorToString error =
    case error of
        Http.BadUrl url ->
            "Invalid URL: " ++ url

        Http.Timeout ->
            "Request timeout"

        Http.NetworkError ->
            "Network error"

        Http.BadStatus status ->
            "Bad status: " ++ String.fromInt status

        Http.BadBody body ->
            "Bad response body: " ++ body


view : Model -> Html Msg
view model =
    let
        header =
            h1 [] [ text "Comments" ]

        comments =
            withSuccess model.lc |> List.map viewComment

        commentsLoading =
            case model.lc of
                Loading ->
                    [ spinner ]

                _ ->
                    []

        form =
            case model.lc of
                Success _ ->
                    [ div []
                        [ h2 [] [ text "Submit a Comment" ]
                        , p [] [ text "Comments will be moderated. Feel free to use markdown/html formatting." ]
                        , viewForm model
                        , viewSubmitStatus model.submitStatus
                        ]
                    ]

                _ ->
                    []

        -- debug = [label [] [ Debug.toString model |> text ]]
        debug =
            []
    in
    div [] (header :: commentsLoading ++ comments ++ form ++ debug)


renderMarkdown : String -> Html msg
renderMarkdown markdownString =
    markdownString
        |> Markdown.parse
        |> Result.mapError (\_ -> "Markdown parsing failed")
        |> Result.andThen
            (\ast ->
                Markdown.Renderer.render Markdown.Renderer.defaultHtmlRenderer ast
                    |> Result.mapError (\e -> "HTML rendering failed: " ++ e)
            )
        |> Result.map (\rendered -> Html.div [] rendered)
        |> Result.withDefault (Html.text markdownString)


renderHtmlContent : String -> Html msg
renderHtmlContent htmlString =
    case Html.Parser.run htmlString of
        Ok nodes ->
            Html.div [] (Html.Parser.Util.toVirtualDom nodes)

        Err _ ->
            Html.text htmlString


withSuccess : State LoadedComments -> List (Comment Time.Posix)
withSuccess s =
    case s of
        Success lc ->
            lc.allComments

        _ ->
            []


viewComment : Comment Time.Posix -> Html Msg
viewComment comment =
    div []
        [ div []
            [ h3 []
                [ text
                    (formatDate comment.timestamp
                        ++ "  - "
                        ++ (if comment.author == "" then
                                "anonymous"

                            else
                                comment.author
                           )
                    )
                ]
            ]
        , div [] [ renderMarkdown comment.commentBody ]
        ]


formatDate : Time.Posix -> String
formatDate timestamp =
    let
        -- No strftime? No Enum? Bounded???
        monthToInt m =
            case m of
                Time.Jan ->
                    1

                Time.Feb ->
                    2

                Time.Mar ->
                    3

                Time.Apr ->
                    4

                Time.May ->
                    5

                Time.Jun ->
                    6

                Time.Jul ->
                    7

                Time.Aug ->
                    8

                Time.Sep ->
                    9

                Time.Oct ->
                    10

                Time.Nov ->
                    11

                Time.Dec ->
                    12

        year =
            Time.toYear Time.utc timestamp |> String.fromInt

        month =
            Time.toMonth Time.utc timestamp |> monthToInt |> String.fromInt |> String.padLeft 2 '0'

        day =
            Time.toDay Time.utc timestamp |> String.fromInt |> String.padLeft 2 '0'

        hour =
            Time.toHour Time.utc timestamp |> String.fromInt |> String.padLeft 2 '0'

        minute =
            Time.toMinute Time.utc timestamp |> String.fromInt |> String.padLeft 2 '0'

        second =
            Time.toSecond Time.utc timestamp |> String.fromInt |> String.padLeft 2 '0'
    in
    year ++ "-" ++ month ++ "-" ++ day ++ " " ++ hour ++ ":" ++ minute ++ ":" ++ second


getComments : Url.Url -> Cmd Msg
getComments url =
    Http.request
        { method = "POST"
        , headers = []
        , url = urlComments
        , body =
            Http.jsonBody <|
                Encode.object
                    [ ( "siteID", Encode.string theSiteId )
                    , ( "postID", Encode.string (trimTrailingSlashes url.path) )
                    , ( "isActive", Encode.bool True )
                    ]
        , expect = Http.expectJson GotComments commentsDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


trimTrailingSlashes : String -> String
trimTrailingSlashes url =
    if String.endsWith "/" url then
        trimTrailingSlashes (String.dropRight 1 url)
    else
        url


commentsDecoder : Decode.Decoder (List (Comment Int))
commentsDecoder =
    Decode.oneOf
        [ Decode.null []
        , Decode.map (List.sortBy .timestamp) (Decode.list commentDecoder)
        , Decode.list commentDecoder
        ]


commentDecoder : Decode.Decoder (Comment Int)
commentDecoder =
    Decode.map8 Comment
        (Decode.field "commentID" Decode.string)
        (Decode.field "siteID" Decode.string)
        (Decode.field "postID" Decode.string)
        (Decode.field "timestamp" Decode.int)
        (Decode.field "author" Decode.string)
        (Decode.field "authorEmail" Decode.string)
        (Decode.field "commentBody" Decode.string)
        (Decode.field "isActive" Decode.bool)


viewForm : Model -> Html Msg
viewForm model =
    form [ onSubmit SubmitForm ]
        [ div []
            [ label [] [ text "Name:" ]
            , div []
                [ input [ Attr.type_ "text", Attr.value model.author, onInput UpdateAuthor ] []
                ]
            ]
        , div []
            [ label [] [ text "Email (optional):" ]
            , div []
                [ input [ Attr.type_ "text", Attr.value model.authorEmail, onInput UpdateAuthorEmail ] []
                ]
            ]
        , div []
            [ label [] [ text "Comment:" ]
            , div []
                [ textarea
                    [ Attr.value model.commentBody
                    , onInput UpdateCommentBody
                    , Attr.style "width" "100%"
                    , Attr.style "height" "450px"
                    ]
                    []
                ]
            ]
        , button
            [ Attr.type_ "submit"
            , Attr.disabled (model.turnstileToken == Nothing)
            ]
            [ text "Submit Comment" ]
        ]


viewSubmitStatus : SubmitStatus -> Html Msg
viewSubmitStatus status =
    case status of
        NotSubmitted ->
            text ""

        -- FIXME timeout?
        Submitting ->
            p [] [ text "Submitting comment..." ]

        SubmitSuccess ->
            p [ Attr.style "color" "green" ] [ text "Comment submitted successfully!" ]

        SubmitFailure error ->
            p [ Attr.style "color" "red" ] [ text ("Error submitting comment: " ++ error) ]
