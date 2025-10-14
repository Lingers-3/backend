function to_env_var(camel_case) {
    result = "";
    len = length(camel_case);

    for (i = 1; i <= len; i++) {
        char = substr(camel_case, i, 1);

        if (i > 1) {
            prev_char = substr(camel_case, i-1, 1);
            next_char = substr(camel_case, i+1, 1);

            if (char ~ /[A-Z]/ && (prev_char ~ /[a-z]/ || prev_char ~ /[0-9]/ || (prev_char ~ /[A-Z]/ && next_char ~ /[a-z]/))) {
                result = result "_";
            }
        }

        result = result toupper(char);
    }

    gsub("__+", "_", result);
    gsub(/^_|_$/, "", result);

    return result;
}

BEGIN {
    print "package config";
    print "import \"os\"";
    print "func Load() *Config {";
    print "return &Config{";
}

/^\s*[a-zA-Z][a-zA-Z0-9]*\s+[a-z]+[a-zA-Z]*$/ {
    if ($1 ~ /^(return|package)$/) { next; }
    printf "%s: os.Getenv(\"%s\"),\n", $1, to_env_var($1);
}

END {
    print "}}";
}
